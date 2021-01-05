package cmd

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var FlagLogLevel = cli.StringFlag{
	Name:     "log-level, l",
	Value:    "info",
	Usage:    "config file name",
	Required: false,
}

var FlagLogFile = cli.StringFlag{
	Name:     "log-file",
	Usage:    "config file name",
	Required: false,
}

func SetLogByContext(context *cli.Context) {

	logLevelStr := context.GlobalString("log-level")

	if logLevelStr != "" {
		logLevel, err := log.ParseLevel(logLevelStr)
		if err != nil {
			log.Fatal("Unable to parse log level", err)
		} else {
			log.SetLevel(logLevel)
		}
	}

	logFileStr := context.GlobalString("log-file")

	if logFileStr != "" {
		file, err := os.OpenFile(logFileStr, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(fmt.Sprintf("Unable to open file [%s]", logFileStr), err)
		} else {
			log.SetOutput(io.MultiWriter(os.Stderr, file))
		}
	}
}
