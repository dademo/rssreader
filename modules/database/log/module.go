package log

import (
	"fmt"
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

	MatchingDataKeys map[string]interface{}

	// Specials
	ScrollID string // Used for the Elasticsearch scroll API in order to retrieve data without submitting the query again and making sure of uniqueness of values
}

type LogEntriesPage struct {
	Entries []LogEntry

	PageNo   uint
	PageSize uint

	TotalElements uint
	TotalPages    uint
}

type LogEntry struct {
	timestamp time.Time
	message   string
	data      *map[string]interface{}
	level     string
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
