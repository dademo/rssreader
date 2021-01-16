package cmd

import "github.com/urfave/cli"

var FlagLog = cli.StringFlag{
	Name:     "log-level",
	Usage:    "Set the logging level",
	Required: false,
	Value:    "info",
}

var GlobalFlags = []cli.StringFlag{
	FlagLog,
	FlagConfig,
}
