package log

import (
	"errors"

	"github.com/dademo/rssreader/modules/log/hook"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDBBackendDefinition struct {
	client *mongo.Client
}

func mongoDBLogBackend() (LogBackend, error) {

	client := hook.GetMongoDBClient()
	if client == nil {
		return nil, errors.New("Unable to get MongoDB client, got nil value")
	}

	return &mongoDBBackendDefinition{
		client: client,
	}, nil
}

func (backendDefinition *mongoDBBackendDefinition) QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error) {
	return nil, nil
}
