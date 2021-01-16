package cmd

import (
	"github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/scheduler"
	"github.com/dademo/rssreader/modules/server"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var CmdServe = cli.Command{
	Name:      "serve",
	ShortName: "s",
	Action:    serve,
}

func serve(context *cli.Context) error {

	SetLogByContext(context)

	appConfig, err := getConfigFromContext(context)

	if err != nil {
		log.Error("Unable to parse configuration")
		return err
	}

	err = database.ConnectDB(appConfig.DbConfig)

	if err != nil {
		log.Error("An error occured when connecting to the database")
		return err
	}

	err = database.PrepareDatabase()

	if err != nil {
		log.Error("An error occured when prepairing the database")
		return err
	}

	jobScheduler := scheduler.New()
	server.ScheduleFromConfig(jobScheduler, appConfig)

	return nil
}
