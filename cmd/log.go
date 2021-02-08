package cmd

import (
	"fmt"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/log"
	logHooks "github.com/dademo/rssreader/modules/log/hook"

	"github.com/sirupsen/logrus"
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

var FlagLogReportCaller = cli.BoolFlag{
	Name:     "log-report-caller",
	Usage:    "write the caller name in the log",
	Required: false,
}

var FlagLogDisableColors = cli.BoolFlag{
	Name:     "log-disable-colors",
	Usage:    "write the caller name in the log",
	Required: false,
}

var FlagLogFullTimestamp = cli.BoolFlag{
	Name:     "log-full-timestamp",
	Usage:    "write the caller name in the log",
	Required: false,
}

func SetLogByContextAndConfig(context *cli.Context, logConfig *config.LogConfig) error {

	logLevelStr := firstNonEmpty(context.GlobalString("log-level"), logConfig.Level)

	if logLevelStr != "" {
		log.SetLogLevel(logLevelStr)
	}

	if context.GlobalBool("log-report-caller") || logConfig.ReportCaller {
		log.SetReportCaller()
	}

	log.SetFormat(&logrus.TextFormatter{
		DisableColors:   context.GlobalBool("log-disable-colors") || logConfig.DisableColors,
		FullTimestamp:   context.GlobalBool("log-full-timestamp") || logConfig.FullTimestamp,
		TimestampFormat: firstNonEmpty(context.GlobalString("log-timestamp-format"), logConfig.TimestampFormat),
	})

	// Workaround to get unique values
	allLogFileStrDict := make(map[string]bool)
	allLogFileStr := make([]string, 0)

	for _, logFileStr := range logConfig.Output {
		allLogFileStrDict[logFileStr] = true
	}
	allLogFileStrDict[context.GlobalString("log-file")] = true

	for logFileStr, _ := range allLogFileStrDict {
		if logFileStr != "" {
			allLogFileStr = append(allLogFileStr, logFileStr)
		}
	}

	// If none set, defaults to stderr
	if len(allLogFileStr) == 0 {
		allLogFileStr = []string{"stderr"}
	}

	err := log.SetLogOutputStreams(allLogFileStr...)
	if err != nil {
		fmt.Println("Unable to set log output streams")
		return err
	}

	return logHooks.RegisterHooks(logConfig.Backends)
}

func SetLogByContext(context *cli.Context) error {
	return SetLogByContextAndConfig(context, config.DefaultLogConfig())
}

func firstNonEmpty(args ...string) string {
	for _, arg := range args {
		if arg != "" {
			return arg
		}
	}
	return ""
}
