package database

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"text/template"

	"github.com/dademo/rssreader/modules/config"

	log "github.com/sirupsen/logrus"
)

type DatabaseTableCreator func(*sql.Conn) error

var (
	registeredTableCreators []DatabaseTableCreator
	database                *sql.DB
	dbDriver                string
)

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

	for _, databaseTableCreator := range registeredTableCreators {
		err = databaseTableCreator(conn)
		if err != nil {
			log.Error("An error occured while creating some tables", err)
			return err
		}
	}

	return
}

func RegisterDatabaseTableCreator(databaseTableCreator DatabaseTableCreator) {
	registeredTableCreators = append(registeredTableCreators, databaseTableCreator)
}

func normalizedSql(sql string) (string, error) {

	var buffer bytes.Buffer

	log.Debug(fmt.Sprintf("Formatting SQL :\n%s", sql))

	tpl, err := template.New("sql").Parse(sql)

	if err != nil {
		return "", err
	}

	err = tpl.Execute(&buffer, struct {
		SqlTimestamp string
	}{
		SqlTimestamp: databaseDateTypes[dbDriver],
	})

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
