package cmd

import (
	"fmt"

	"github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/feed"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var CmdRun = cli.Command{
	Name:      "run",
	ShortName: "r",
	Action:    run,
}

func run(cliContext *cli.Context) error {

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
		log.WithError(err).Error("An error occured while connecting to the database")
		return err
	}

	err = database.PrepareDatabase()

	if err != nil {
		log.WithError(err).Error("An error occured while prepairing the database")
		return err
	}

	fetchedFeeds, err := feed.FetchAll(appConfig)
	if err != nil {
		log.WithError(err).Error("Unable to fetch feeds")
		return err
	}

	for _, fetchedFeed := range fetchedFeeds {
		err = fetchedFeed.Save()
		if err != nil {
			log.Debug(fmt.Sprintf("An error occured while saving feed [%s]", fetchedFeed.Title))
			return err
		}
	}

	log.Info("Feeds saved")

	return nil
}
