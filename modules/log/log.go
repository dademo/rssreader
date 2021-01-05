package log

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	specialValues = map[string]io.Writer{
		"stdout": os.Stdout,
		"stderr": os.Stderr,
	}
)

func SetLogOutputStreams(streamStrs ...string) error {

	streams := make([]io.Writer, 0, len(streamStrs))

	for _, streamStr := range streamStrs {

		stream, err := openStream(streamStr)

		if err != nil {
			log.Error(fmt.Sprintf("Error while opening stream [%s]", streamStr))
			return err
		}

		streams = append(streams, stream)
	}

	log.SetOutput(io.MultiWriter(streams...))

	return nil
}

func SetLogLevel(logLevelStr string) error {

	logLevel, err := log.ParseLevel(logLevelStr)

	if err != nil {
		log.Error("Unable to parse log level")
		return err
	} else {
		log.SetLevel(logLevel)
		return nil
	}
}

func openStream(streamStr string) (io.Writer, error) {

	if val, ok := specialValues[streamStr]; ok {
		return val, nil
	} else {
		return os.OpenFile(streamStr, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
}
