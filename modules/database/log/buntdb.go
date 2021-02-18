package log

import (
	"encoding/json"
	"errors"
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
	logLevel      logrus.Level
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

			if comparator.compareEntryKey(key) {

				// First pages we want to skip
				if it < query.Page.PageNo*query.Page.PageSize {
					it++
					return true
				}

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
				it++
				pageResult.PageSize++

				return it <= (query.Page.PageNo+1)*query.Page.PageSize
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

func (comparator *buntDbComparator) compareEntryKey(key string) bool {

	keyItems := comparator.keyComparator.FindStringSubmatch(key)

	if keyItems == nil {
		//log.Debug(fmt.Sprintf("Not match for key [%s]", key))
		return false
	}
	if len(keyItems) != 5 {
		//log.Debug(fmt.Sprintf("Unable to fetch keys from key, expected 5, got %d [%s]", len(keyItems), key))
		return false
	}

	timestamp, err := strconv.ParseInt(keyItems[1], 10, 0)
	if err != nil {
		log.DebugError(err, "Unable to parse timestamp")
		return false
	}

	if !comparator.compareTimestamp(time.Unix(0, timestamp)) {
		return false
	}

	match, err := comparator.compareLogLevel(keyItems[2])
	if err != nil {
		log.DebugError(err, "Unable to compare log level")
		return false
	}
	if !match {
		return false
	}

	match, err = comparator.compareFile(keyItems[3])
	if err != nil {
		log.DebugError(err, "Unable to compare file")
		return false
	}
	if !match {
		return false
	}

	match, err = comparator.compareFunction(keyItems[4])
	if err != nil {
		log.DebugError(err, "Unable to compare function")
		return false
	}
	if !match {
		return false
	}

	return true
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

func (comparator *buntDbComparator) compareLogLevel(logLevel string) (bool, error) {

	// Levels are set from the less to the most precise
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return false, err
	}

	return compare(
		int64(level-comparator.logLevel),
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
