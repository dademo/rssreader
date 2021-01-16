package database

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"

	//_ "github.com/dademo/rssreader/modules/database/feed"

	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type DatabaseModuleTableCreator = func(connection *sql.Tx) error
type DatabaseModuleTableUpdater = func(connection *sql.Tx, oldVersion string) error
type DatabaseModuleOnDatabaseSetFct = func(db *sql.DB)

type DatabaseModuleDescption struct {
	ModuleName string
	Version    string
}

type DatabaseModuleTableCreationDef struct {
	ModuleName                 string
	Version                    string
	DatabaseModuleTableCreator DatabaseModuleTableCreator
	DatabaseModuleTableUpdater DatabaseModuleTableUpdater
}

var (
	registeredTableCreatorDefs         []DatabaseModuleTableCreationDef
	registeredModulesOnDatabaseSetFcts []DatabaseModuleOnDatabaseSetFct
	database                           *sql.DB
	dbDriver                           string
)

var databasePrimaryKeyTypes = map[string]string{
	"sqlite":     "INTEGER PRIMARY KEY NOT NULL",
	"sqlite3":    "INTEGER PRIMARY KEY NOT NULL",
	"mysql":      "INTEGER PRIMARY KEY NOT NULL AUTO_INCREMENT",
	"postgres":   "SERIAL PRIMARY KEY NOT NULL",
	"postgresql": "SERIAL PRIMARY KEY NOT NULL",
}

var databaseDateTypes = map[string]string{
	"sqlite":     "TEXT",
	"sqlite3":    "TEXT",
	"mysql":      "DATETIME",
	"postgres":   "TIMESTAMP WITHOUT TIMEZONE",
	"postgresql": "TIMESTAMP WITHOUT TIMEZONE",
}

type DatabaseEntity interface {
	Save() error
	Refresh() error
}

type PrimaryKey = uint64

func ConnectDB(dbConfig config.DatabaseConfig) error {

	var err error
	log.Debug(fmt.Sprintf("Connecting to the database with driver [%s] with connection string [%s]", dbConfig.Driver, dbConfig.ConnStr))
	database, err = sql.Open(dbConfig.Driver, dbConfig.ConnStr)

	if err != nil {
		return err
	}

	dbDriver = dbConfig.Driver
	log.Debug("Connected to the database")

	err = database.Ping()
	if err != nil {
		log.Error("Unable to establish a connection to the database")
		return err
	}

	return nil
}

func Cleanup() {

	if database != nil {
		database.Close()
		database = nil
	}
}

func PrepareDatabase() (err error) {

	log.Debug("Prepairing database")

	ctx := context.Background()
	defer ctx.Done()
	conn, err := database.Conn(ctx)
	if err != nil {
		log.Error("Unable to connect to the database", err)
		return err
	}

	defer func() {
		log.Debug("Closing connection")
		cerr := conn.Close()
		if err == nil {
			err = cerr
		}
		log.Debug("Connection closed")
	}()

	tx, err := database.Begin()
	if err != nil {
		log.Debug("Unable to begin transaction")
		return err
	}

	err = initInformationTables(tx)
	if err != nil {
		log.Debug("Unable to initalize system tables")
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Debug("Unable to commit transaction")
		return err
	}

	for _, databaseTableCreatorDef := range registeredTableCreatorDefs {

		tx, err = database.Begin()
		if err != nil {
			log.Debug("Unable to begin transaction")
			return err
		}

		existingModuleDef, err := fetchModuleByName(tx, databaseTableCreatorDef.ModuleName)

		if err != nil {
			log.Error("Unable to check for module installation, ", err)
			return err
		}

		if existingModuleDef == nil {
			log.Debug(fmt.Sprintf("Creating tables for mod [%s:%s]", databaseTableCreatorDef.ModuleName, databaseTableCreatorDef.Version))
			err = databaseTableCreatorDef.DatabaseModuleTableCreator(tx)
			if err != nil {
				log.Error("An error occured while creating some tables, ", err)
				return err
			}
			log.Debug("Table created")

			log.Debug("Saving the new module status")
			err = saveModuleByName(
				tx,
				DatabaseModuleDescption{
					ModuleName: databaseTableCreatorDef.ModuleName,
					Version:    databaseTableCreatorDef.Version,
				},
				false,
			)
			if err != nil {
				log.Debug("Unable to save the new module status")
				return err
			}
			log.Debug("Module status saved")
		} else if existingModuleDef != nil && existingModuleDef.Version != databaseTableCreatorDef.Version {
			log.Debug(fmt.Sprintf("Updating tables for mod [%s:%s] from version [%s]", databaseTableCreatorDef.ModuleName, databaseTableCreatorDef.Version, existingModuleDef.Version))
			err = databaseTableCreatorDef.DatabaseModuleTableUpdater(tx, existingModuleDef.Version)
			if err != nil {
				log.Error("An error occured while updating some tables, ", err)
				return err
			}
			log.Debug("Table updated")

			log.Debug("Saving the new module status")
			err = saveModuleByName(
				tx,
				DatabaseModuleDescption{
					ModuleName: databaseTableCreatorDef.ModuleName,
					Version:    databaseTableCreatorDef.Version,
				},
				true,
			)
			if err != nil {
				log.Debug("Unable to save the new module status")
				return err
			}
			log.Debug("Module status saved")
		} else {
			log.Debug(fmt.Sprintf("Nothing to do for mod [%s:%s]", databaseTableCreatorDef.ModuleName, databaseTableCreatorDef.Version))
		}
		err = tx.Commit()
		if err != nil {
			log.Debug("Unable to commit transaction")
			return err
		}
	}

	for _, dbRegisteredFct := range registeredModulesOnDatabaseSetFcts {
		dbRegisteredFct(database)
	}

	return
}

func RegisterDatabaseTableCreator(databaseTableCreatorDef DatabaseModuleTableCreationDef) {
	registeredTableCreatorDefs = append(registeredTableCreatorDefs, databaseTableCreatorDef)
}

func RegisterOnDatabaseSet(onDatabaseSetRef DatabaseModuleOnDatabaseSetFct) {
	registeredModulesOnDatabaseSetFcts = append(registeredModulesOnDatabaseSetFcts, onDatabaseSetRef)
}

func NormalizedSql(sql string) (string, error) {

	var buffer bytes.Buffer

	log.Debug(fmt.Sprintf("Formatting SQL :\n%s", sql))

	tpl, err := template.New("sql").Parse(sql)

	if err != nil {
		return "", err
	}

	err = tpl.Execute(&buffer, struct {
		SqlPrimaryKey string
		SqlTimestamp  string
	}{
		SqlPrimaryKey: databasePrimaryKeyTypes[dbDriver],
		SqlTimestamp:  databaseDateTypes[dbDriver],
	})

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func EntityId(e interface{}) interface{} {

	const fieldIdName = "Id"

	reflected := reflect.ValueOf(e)

	if !reflected.IsNil() {

		if reflected.Type().Kind() != reflect.Ptr {
			reflected = reflect.New(reflect.TypeOf(e))
		}

		reflectedFieldId := reflected.Elem().FieldByName(fieldIdName)

		if reflectedFieldId.IsValid() {

			switch reflectedFieldId.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fieldRealValue := reflectedFieldId.Int()
				if fieldRealValue != 0 {
					return fieldRealValue
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fieldRealValue := reflectedFieldId.Uint()
				if fieldRealValue != 0 {
					return fieldRealValue
				}
			}
		}
	}
	return nil
}

func SqlExec(statement *sql.Stmt, args ...interface{}) (int64, error) {

	result, err := statement.Exec(args...)
	if err != nil {
		log.Debug("An error occured while running a statement")
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Debug("Unable to get affected rows count")
	} else {
		log.Debug(fmt.Sprintf("%d rows affected", rowsAffected))
	}

	_id, err := result.LastInsertId()
	if err != nil {
		log.Debug(fmt.Sprintf("An error occured while getting the last inserted Id"))
		return 0, err
	} else {
		return _id, nil
	}
}

func SqlDateParse(value interface{}) (*time.Time, error) {

	if value == nil {
		return nil, nil
	}

	vTime, ok := value.(time.Time)
	if ok {
		return &vTime, nil
	}

	vString, ok := value.(string)
	if ok {
		switch dbDriver {
		case "sqlite", "sqlite3":
			for _, dateFormat := range sqlite3.SQLiteTimestampFormats {
				date, err := time.Parse(dateFormat, vString)
				if err != nil {
					log.Debug(fmt.Sprintf("Unable to parse date using format [%s]", dateFormat))
				} else {
					return &date, nil
				}
			}
			return nil, errors.New(fmt.Sprintf("Unable to find a suitable date format for value [%s]", vString))
		default:
			return nil, errors.New("String given while driver supports dates")
		}
	}

	return nil, errors.New(fmt.Sprintf("Unable to parse sql date (unknown type for value [%s] of type [%s])", value, reflect.TypeOf(value)))
}

func initInformationTables(connection *sql.Tx) error {

	sql := `
		CREATE TABLE IF NOT EXISTS system_information_table (
			id 			{{.SqlPrimaryKey}},
			module 		TEXT NOT NULL UNIQUE,
			version 	TEXT NOT NULL,
			last_update	{{.SqlTimestamp}}
		);
	`

	log.Debug("Creating system tables")

	sql, err := NormalizedSql(sql)

	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("Running command :\n%s", sql))

	_, err = connection.Exec(sql)

	if err != nil {
		log.Error("Unable to create sytsem tables")
		return err
	}

	log.Debug("System tables created")
	return nil
}

func fetchModuleByName(tx *sql.Tx, moduleName string) (*DatabaseModuleDescption, error) {

	sql := `
		SELECT
			module,
			version
		FROM system_information_table
		WHERE module = ?
	`
	v := new(DatabaseModuleDescption)

	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Debug("Unable to prepare statement")
		return nil, err
	}
	defer DeferStmtCloseFct(stmt)

	rows, err := stmt.Query(moduleName)
	if err != nil {
		log.Debug("Unable to get result row")
		return nil, err
	}
	if rows.Err() != nil {
		log.Debug("Unable to get result row")
		return nil, rows.Err()
	}
	defer DeferRowsCloseFct(rows)

	if rows.Next() {
		err = rows.Scan(&v.ModuleName, &v.Version)
		if err != nil {
			log.Debug("Unable to affect results")
			return nil, err
		} else {
			return v, nil
		}
	} else {
		return nil, nil
	}
}

func saveModuleByName(tx *sql.Tx, databaseModuleDescription DatabaseModuleDescption, update bool) error {

	var sql string
	if update {
		sql = `
			UPDATE system_information_table SET
				version = ?,
				last_update = ?
			WHERE module = ?
		`
	} else {
		sql = `
			INSERT INTO system_information_table(version, last_update, module)
			VALUES (?, ?, ?)
		`
	}

	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Debug("Unable to create the statement for database version creation or update")
		return err
	}
	defer stmt.Close()

	_, err = SqlExec(stmt,
		databaseModuleDescription.Version,
		time.Now(),
		databaseModuleDescription.ModuleName,
	)

	if err != nil {
		log.Debug("An error occured while adding or updating a module version")
		return err
	}

	return nil
}

func DeferStmtCloseFct(stmt *sql.Stmt) func() {
	return func() {
		err := stmt.Close()
		if err != nil {
			appLog.DebugError("Unable to close statement, ", err)
		}
	}
}

func DeferRowsCloseFct(rows *sql.Rows) func() {
	return func() {
		err := rows.Close()
		if err != nil {
			appLog.DebugError("Unable to close rows, ", err)
		}
	}
}
