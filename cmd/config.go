package cmd

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var FlagConfig = cli.StringFlag{
	Name:      "config, c",
	Usage:     "config file to use",
	TakesFile: true,
	Required:  true,
}

var CmdConfig = cli.Command{
	Name:      "checkConfig",
	ShortName: "check, c",
	Action:    check,
}

func check(cliContext *cli.Context) error {

	SetLogByContext(cliContext)

	_, err := getConfigFromContext(cliContext)

	if err != nil {
		log.WithError(err).Error("Unable to parse configuration")
		return err
	} else {
		fmt.Println("Your configuration is correct")
		return nil
	}
}

func getConfigFromContext(context *cli.Context) (*config.Config, error) {
	return config.ReadConfig(context.GlobalString("config"))
}
