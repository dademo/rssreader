package cmd

import (
	"net/http"

	"github.com/dademo/rssreader/modules/database"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/scheduler"
	"github.com/dademo/rssreader/modules/server"
	"github.com/dademo/rssreader/modules/web"

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
	log.Debug("Database initialized")

	log.Debug("Prepairing http server")
	jobScheduler := scheduler.New()
	server.ScheduleFromConfig(jobScheduler, appConfig)

	httpServeMux := http.NewServeMux()
	err = web.RegisterServerHandlers(httpServeMux, appConfig.HttpConfig)

	if err != nil {
		appLog.DebugError("Unable to register routes")
		return err
	}

	web.SetDisplayErrors(appConfig.HttpConfig.DisplayErrors)

	log.Debug("Serving http")
	return http.ListenAndServe(appConfig.HttpConfig.ListenAddress, httpServeMux)
}
