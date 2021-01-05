package database

import (
	"context"
	"database/sql"

	log "github.com/sirupsen/logrus"
)

type DatabaseTableCreator func(*sql.Conn) error

var (
	registeredTableCreators []DatabaseTableCreator
	ctx                     context.Context
)

func PrepareDatabase(database *sql.DB) (err error) {

	conn, err := database.Conn(ctx)
	if err != nil {
		log.Error("Unable to connect to the database", err)
		return err
	}

	defer func() {
		cerr := conn.Close()
		if err == nil {
			err = cerr
		}
	}()

	for _, databaseTableCreator := range registeredTableCreators {
		err = databaseTableCreator(conn)
		if err != nil {
			log.Error("An error occured while creating table", err)
			return err
		}
	}

	return
}

func RegisterDatabaseTableCreator(databaseTableCreator DatabaseTableCreator) {
	registeredTableCreators = append(registeredTableCreators, databaseTableCreator)
}
