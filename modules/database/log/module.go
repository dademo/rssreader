package log

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/log/hook"
)

var backends = map[string]LogBackendGetter{
	hook.BackendNameElasticsearch: elasticsearchLogBackend,
	hook.BackendNameInfluxDB:      influxDBLogBackend,
	hook.BackendNameMongoDB:       mongoDBLogBackend,
	hook.BackendNameBuntDB:        buntDBLogBackend,
}

type LogBackendGetter func() (LogBackend, error)

type LogQueryOpts struct {
	Page database.PageQuery

	Level           string
	LevelComparator database.Comparator

	Date           time.Time
	DateComparator database.Comparator

	Message           string
	MessageComparator database.StringComparator

	File           string
	FileComparator database.StringComparator

	Function           string
	FunctionComparator database.StringComparator

	MatchingDataKeys map[string]interface{}

	// Specials
	ScrollID string // Used for the Elasticsearch scroll API in order to retrieve data without submitting the query again and making sure of uniqueness of values
}

type LogEntriesPage struct {
	Entries []*LogEntry

	PageNo   uint
	PageSize uint

	TotalElements uint
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	File      string
	Function  string
	Message   string
	Data      map[string]interface{}
}

type LogBackend interface {
	QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error)
}

func GetLogBackendFromString(backendName string) (LogBackend, error) {

	if val, ok := backends[strings.ToLower(backendName)]; ok {
		return val()
	} else {
		return nil, fmt.Errorf("Unable to locate backend named [%s]", backendName)
	}
}

func compare(compareValue int64, comparator database.Comparator) bool {

	switch comparator {
	default:
		fallthrough
	case database.ComparatorGreaterEqualThan:
		return compareValue >= 0
	case database.ComparatorGreaterThan:
		return compareValue > 0
	case database.ComparatorEquals:
		return compareValue == 0
	case database.ComparatorLowerThan:
		return compareValue < 0
	case database.ComparatorLowerEqualThan:
		return compareValue <= 0
	}
}

func compareStr(value string, compareTo string, comparator database.StringComparator) (bool, error) {

	switch comparator {
	default:
		fallthrough
	case database.StrComparatorEquals:
		return value == compareTo, nil
	case database.StrComparatorContains:
		return strings.Contains(value, compareTo), nil
	case database.StrComparatorMatches:
		return regexp.MatchString(compareTo, value)
	}
}
