package main

import (
	"os"
	"sort"

	"github.com/dademo/rssreader/cmd"
	"github.com/dademo/rssreader/modules/database"

	appLog "github.com/dademo/rssreader/modules/log"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	appVersion = "development"
)

func main() {

	app := cli.NewApp()
	app.Name = "RSS reader"
	app.Usage = "A rss stream sync service"
	app.Description = "An application syncing your rss streams"
	app.Version = appVersion

	app.Flags = []cli.Flag{
		cmd.FlagLogLevel,
		cmd.FlagLogFile,
		cmd.FlagConfig,
	}

	app.Commands = []cli.Command{
		cmd.CmdRun,
		cmd.CmdConfig,
		cmd.CmdServe,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}
	cleanup()
}

func cleanup() {
	database.Cleanup()
	appLog.Cleanup()
}
