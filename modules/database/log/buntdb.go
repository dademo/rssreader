package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/log/hook"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
)

type buntDBBackendDefinition struct {
	db *buntdb.DB
}

type buntDbComparator struct {
	keyComparator *regexp.Regexp
	query         *LogQueryOpts
}

func buntDBLogBackend() (LogBackend, error) {

	db := hook.GetBuntClient()
	if db == nil {
		return nil, errors.New("Unable to get BuntDB client, got nil value")
	}

	return &buntDBBackendDefinition{
		db: db,
	}, nil
}

func (backendDefinition *buntDBBackendDefinition) QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error) {

	//var expectedLogLevel logrus.Level
	var err error

	pageResult := &LogEntriesPage{
		Entries:       make([]*LogEntry, 0, query.Page.PageSize),
		PageNo:        query.Page.PageNo,
		PageSize:      0,
		TotalElements: 0,
	}

	keyComparator, err := regexp.Compile("log:([0-9]+):(.+):(.+):(.+)")
	if err != nil {
		log.DebugError(err, "Unable to crete bunt key comparator regular expression")
		return nil, err
	}

	comparator := buntDbComparator{
		keyComparator: keyComparator,
		query:         query,
	}

	err = backendDefinition.db.View(func(tx *buntdb.Tx) error {

		var internalError error
		it := uint(0)

		ascendError := tx.Ascend(hook.IndexBuntDB, func(key string, value string) bool {

			var logEntry LogEntry
			var keyMatches bool

			keyMatches, internalError = comparator.compareEntryKey(key)

			if internalError != nil {
				return false
			}

			if keyMatches {

				// We only generate wanted page
				if it >= query.Page.PageNo*query.Page.PageSize &&
					it < (query.Page.PageNo+1)*query.Page.PageSize {

					var buntDBLogEntry hook.BuntDBLogEntry

					internalError = json.Unmarshal([]byte(value), &buntDBLogEntry)

					if internalError != nil {
						return false
					}

					logEntry.Timestamp = time.Unix(0, buntDBLogEntry.Timestamp)
					logEntry.Level = buntDBLogEntry.Level
					logEntry.Message = buntDBLogEntry.Message
					logEntry.File = buntDBLogEntry.File
					logEntry.Function = buntDBLogEntry.Function
					logEntry.Timestamp = time.Unix(0, buntDBLogEntry.Timestamp)
					logEntry.Data = buntDBLogEntry.Data

					pageResult.Entries = append(pageResult.Entries, &logEntry)
					pageResult.PageSize++
				}
				it++
				pageResult.TotalElements++

				return true
			}
			return true
		})

		if internalError != nil {
			return internalError
		}
		if ascendError != nil {
			return ascendError
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pageResult, nil
}

func (comparator *buntDbComparator) compareEntryKey(key string) (bool, error) {

	keyItems := comparator.keyComparator.FindStringSubmatch(key)

	if keyItems == nil {
		return false, fmt.Errorf("Not match for key [%s]", key)
	}
	if len(keyItems) != 5 {
		return false, fmt.Errorf("Unable to fetch keys from key, expected 5, got %d [%s]", len(keyItems), key)
	}

	timestamp, err := strconv.ParseInt(keyItems[1], 10, 0)
	if err != nil {
		return false, errors.New("Unable to parse timestamp")
	}

	if !comparator.compareTimestamp(time.Unix(0, timestamp)) {
		return false, nil
	}

	match, err := comparator.compareLogLevel(keyItems[2])
	if err != nil {
		return false, fmt.Errorf("Unable to compare log level, %s", err)
	}
	if !match {
		return false, nil
	}

	match, err = comparator.compareFile(keyItems[3])
	if err != nil {
		return false, errors.New("Unable to compare file")
	}
	if !match {
		return false, nil
	}

	match, err = comparator.compareFunction(keyItems[4])
	if err != nil {
		return false, errors.New("Unable to compare function")
	}
	if !match {
		return false, nil
	}

	return true, nil
}

func (comparator *buntDbComparator) compareTimestamp(timestamp time.Time) bool {

	if comparator.query.Date != (time.Time{}) {
		return compare(
			timestamp.UnixNano()-comparator.query.Date.UnixNano(),
			comparator.query.DateComparator,
		)
	}
	return true
}

func (comparator *buntDbComparator) compareLogLevel(keyLogLevelStr string) (bool, error) {

	// Levels are set from the less to the most precise
	keyLogLevel, err := logrus.ParseLevel(keyLogLevelStr)
	if err != nil {
		return false, err
	}

	queryLogLevel, err := logrus.ParseLevel(comparator.query.Level)
	if err != nil {
		return false, err
	}

	return compare(
		int64(queryLogLevel)-int64(keyLogLevel),
		comparator.query.LevelComparator,
	), nil
}

func (comparator *buntDbComparator) compareFile(file string) (bool, error) {

	if comparator.query.File != "" {
		return compareStr(
			file,
			comparator.query.File,
			comparator.query.FileComparator,
		)
	}
	return true, nil
}

func (comparator *buntDbComparator) compareFunction(function string) (bool, error) {

	if comparator.query.Function != "" {
		return compareStr(
			function,
			comparator.query.Function,
			comparator.query.FunctionComparator,
		)
	}
	return true, nil
}
