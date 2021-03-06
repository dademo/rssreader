package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

const shutdownTimeoutMilliseconds = 5000

func serve(cliContext *cli.Context) error {

	appConfig, err := getConfigFromContext(cliContext)

	if err != nil {
		log.WithError(err).Error("Unable to parse configuration")
		return err
	}

	err = SetLogByContextAndConfig(cliContext, appConfig.LogConfig)
	if err != nil {
		log.WithError(err).Error("Unable to set log configuration")
		return err
	}

	err = database.ConnectDB(appConfig.DbConfig)

	if err != nil {
		log.WithError(err).Error("An error occured when connecting to the database")
		return err
	}

	err = database.PrepareDatabase()

	if err != nil {
		log.WithError(err).Error("An error occured when prepairing the database")
		return err
	}
	log.Debug("Database initialized")

	log.Debug("Prepairing http server")
	jobScheduler := scheduler.New()
	server.ScheduleFromConfig(jobScheduler, appConfig)

	httpServeMux := http.NewServeMux()
	err = web.RegisterServerHandlers(httpServeMux, appConfig.HttpConfig)

	if err != nil {
		appLog.DebugError(err, "Unable to register routes")
		return err
	}

	web.SetDisplayErrors(appConfig.HttpConfig.DisplayErrors)

	log.Debug("Creating server")
	srv := http.Server{
		Addr:    appConfig.HttpConfig.ListenAddress,
		Handler: httpServeMux,
	}

	log.Debug("Registering signal handlers")
	var wait sync.WaitGroup
	sigChan := make(chan os.Signal)
	quit := make(chan int)

	wait.Add(1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				log.Debug(fmt.Sprintf("Received %s signal", sig.String()))

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeoutMilliseconds*time.Millisecond))
				defer cancel()

				err = srv.Shutdown(ctx)
				if err != nil {
					appLog.DebugError(err, "An error occured on http server shutown, ", err)
				}

				wait.Done()
				return
			case <-quit:
				wait.Done()
				return
			}
		}

	}()

	log.Debug(fmt.Sprintf("Serving on %s", appConfig.HttpConfig.ListenAddress))
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		quit <- 0
		return err
	}
	log.Debug("Server closed")

	wait.Wait()
	return nil
}
