package log

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

var (
	specialStreamValues = map[string]io.WriteCloser{
		"stdout": os.Stdout,
		"stderr": os.Stderr,
	}
	sysLogStreamValue = "syslog"
	openedStreams     []io.WriteCloser
	defaultLogger     *logrus.Logger
)

func SetLogOutputStreams(streamStrs ...string) error {

	cleanOpenedStreams()

	streams := make([]io.Writer, 0, len(streamStrs))

	for _, streamStr := range streamStrs {

		if streamStr == sysLogStreamValue {
			hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
			if err != nil {
				logrus.WithError(err).Error("Unable to attach syslog, ", err)
			} else {
				logrus.AddHook(hook)
			}
		} else {
			stream, err := openStream(streamStr)

			if err != nil {
				logrus.WithError(err).Error(fmt.Sprintf("Error while opening stream [%s]", streamStr))
				return err
			}

			streams = append(streams, stream)
		}
	}

	logrus.SetOutput(io.MultiWriter(streams...))

	return nil
}

func SetLogLevel(logLevelStr string) error {

	logLevel, err := logrus.ParseLevel(logLevelStr)

	if err != nil {
		logrus.WithError(err).Error("Unable to parse log level")
		return err
	} else {
		logrus.SetLevel(logLevel)
		return nil
	}
}

func SetReportCaller() {
	logrus.SetReportCaller(true)
}

func SetFormat(format *logrus.TextFormatter) {
	logrus.SetFormatter(format)
}

func Cleanup() {
	cleanOpenedStreams()
}

func cleanOpenedStreams() {

	logrus.Debug("Closing opened streams")

	for _, stream := range openedStreams {

		if stream != os.Stderr && stream != os.Stdout {

			err := stream.Close()

			if err != nil {
				logrus.WithError(err).Error()
			}
		}
	}
}

func Debug(args ...interface{}) {
	DebugError(nil, args...)
}

func DebugError(err error, args ...interface{}) {

	args = append(args, ", ", string(debug.Stack()))

	if err != nil {
		logrus.WithError(err).Debug(args...)
	} else {
		logrus.Debug(args...)
	}
}

func LoggerFallback() *logrus.Logger {
	return getDefaultLogrusLogger()
}

func openStream(streamStr string) (io.WriteCloser, error) {

	if val, ok := specialStreamValues[streamStr]; ok {
		return val, nil
	} else {
		stream, err := os.OpenFile(streamStr, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err == nil {
			openedStreams = append(openedStreams, stream)
		}
		return stream, err
	}
}

func getDefaultLogrusLogger() *logrus.Logger {

	if defaultLogger == nil {

		logger := logrus.New()

		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
		})
		logger.SetLevel(logrus.InfoLevel)
		logger.SetOutput(os.Stderr)
		defaultLogger = logger
	}

	return defaultLogger
}
