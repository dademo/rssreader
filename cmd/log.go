package cmd

import (
	"github.com/dademo/rssreader/modules/log"
	"github.com/urfave/cli"
)

var FlagLogLevel = cli.StringFlag{
	Name:     "log-level, l",
	Value:    "info",
	Usage:    "log level",
	Required: false,
}

var FlagLogFile = cli.StringFlag{
	Name:     "log-file",
	Usage:    "file where to write logs",
	Required: false,
}

func SetLogByContext(context *cli.Context) error {

	logLevelStr := context.GlobalString("log-level")

	if logLevelStr != "" {
		log.SetLogLevel(logLevelStr)
	}

	logFileStr := context.GlobalString("log-file")

	if logFileStr != "" {
		return log.SetLogOutputStreams("stderr", logFileStr)
	}
	return nil
}
