package cmd

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var FlagConfig = cli.StringFlag{
	Name:      "config, c",
	Usage:     "config file name",
	TakesFile: true,
	Required:  true,
}

var CmdConfig = cli.Command{
	Name:      "checkConfig",
	ShortName: "check",
	Category:  "Test",
	Action:    check,
}

func check(context *cli.Context) error {
	SetLogByContext(context)
	_, err := config.ReadConfig(context.GlobalString("config"))
	if err != nil {
		log.Fatal("Unable to parse configuration")
		return err
	} else {
		fmt.Println("Your configuration is correct")
		return nil
	}
}
