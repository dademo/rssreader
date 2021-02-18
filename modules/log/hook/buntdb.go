package hook

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

type BuntDBLogHook struct {
	config *config.BuntDBLogBackendConfig
	db     *buntdb.DB
	levels []logrus.Level
}

type BuntDBLogEntry struct {
	Timestamp int64         `json:"timestamp,omitempty"`
	Level     string        `json:"level,omitempty"`
	File      string        `json:"file,omitempty"`
	Function  string        `json:"func,omitempty"`
	Message   string        `json:"message,omitempty"`
	Data      logrus.Fields `json:"data,omitempty"`
}

var buntDBClient *buntdb.DB

const IndexBuntDB = "log"

func connectBuntDB(config *config.BuntDBLogBackendConfig) (*buntdb.DB, error) {

	db, err := buntdb.Open(config.File)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error(fmt.Sprintf("Unable to open BuntDB file [%s]", config.File))
		return nil, err
	}

	logrus.RegisterExitHandler(func() {
		closeErr := db.Close()
		if closeErr != nil {
			appLog.LoggerFallback().WithError(closeErr).Error("An error occured when closing BuntDB database")
		}
	})

	buntDBClient = db

	db.CreateIndex(IndexBuntDB, "log:*", buntdb.IndexString)

	return db, nil
}

func (hook BuntDBLogHook) Fire(entry *logrus.Entry) error {

	buntBody, err := makeBuntBody(hook.config, entry)
	if err != nil {
		return err
	}

	hook.db.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(
			makeBuntKey(entry),
			buntBody,
			makeBuntOption(hook.config.ExpirationSeconds),
		)
		if err != nil {
			appLog.LoggerFallback().Debug("An error occured when saving a log in BuntDB")
			return err
		}
		return nil
	})

	return nil
}

func (hook BuntDBLogHook) Levels() []logrus.Level {
	return hook.levels
}

func makeBuntKey(entry *logrus.Entry) string {

	var file string
	var function string
	if entry.HasCaller() {
		file = strings.ReplaceAll(entry.Caller.File, ":", "_")
		function = strings.ReplaceAll(entry.Caller.Function, ":", "_")
	} else {
		file = "-"
		function = "-"
	}

	return fmt.Sprintf("log:%d:%s:%s:%s",
		entry.Time.UTC().UnixNano(),
		entry.Level,
		file,
		function,
	)
}

func makeBuntBody(config *config.BuntDBLogBackendConfig, entry *logrus.Entry) (string, error) {

	var file string
	var function string
	if entry.HasCaller() {
		file = entry.Caller.File
		function = entry.Caller.Function
	}

	marshalledValue, err := json.Marshal(BuntDBLogEntry{
		Timestamp: entry.Time.UnixNano(),
		Level:     strings.ToUpper(entry.Level.String()),
		File:      file,
		Function:  function,
		Message:   entry.Message,
		Data:      mergeAdditionalTags(config.AdditionalTags, entry.Data),
	})
	if err != nil {
		appLog.LoggerFallback().Debug("Unable to marshall BuntDB log")
		return "", err
	}

	return string(marshalledValue), nil
}

func makeBuntOption(expirationSeconds int) *buntdb.SetOptions {
	if expirationSeconds != 0 {
		return &buntdb.SetOptions{
			Expires: true,
			TTL:     time.Duration(expirationSeconds * int(time.Second)),
		}
	} else {
		return nil
	}
}

func GetBuntClient() *buntdb.DB {
	return buntDBClient
}
