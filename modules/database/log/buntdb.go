package log

import (
	"errors"
	"fmt"

	"github.com/dademo/rssreader/modules/log/hook"
	"github.com/tidwall/buntdb"
)

type buntDBBackendDefinition struct {
	db *buntdb.DB
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
	fmt.Println("Bunt Backend !")
	fmt.Println(query)
	fmt.Println(query.Page)
	return nil, nil
}
