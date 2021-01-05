package cmd

import "github.com/urfave/cli"

var LogFlag = cli.StringFlag{
	Name:     "log-level",
	Usage:    "Set the logging level",
	Required: false,
	Value:    "info",
}
