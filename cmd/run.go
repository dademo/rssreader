package cmd

import (
	"github.com/dademo/rssreader/modules/database"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var CmdRun = cli.Command{
	Name:      "run",
	ShortName: "r",
	Action:    run,
}

func run(context *cli.Context) error {

	SetLogByContext(context)

	config, err := getConfigFromContext(context)

	if err != nil {
		log.Error("Unable to parse configuration")
		return err
	}

	err = database.ConnectDB(config.DbConfig)

	if err != nil {
		log.Error("An error occured when connecting to the database")
		return err
	}

	err = database.PrepareDatabase()

	if err != nil {
		log.Error("An error occured when prepairing the database")
		return err
	}

	return nil
}
