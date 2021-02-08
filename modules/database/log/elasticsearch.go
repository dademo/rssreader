package log

import (
	"errors"

	"github.com/dademo/rssreader/modules/log/hook"
	"github.com/elastic/go-elasticsearch/v7"
)

type elasticsearchBackendDefinition struct {
	client *elasticsearch.Client
}

func elasticsearchLogBackend() (LogBackend, error) {

	client := hook.GetElasticsearchClient()
	if client == nil {
		return nil, errors.New("Unable to get Elasticsearch client, got nil value")
	}

	return &elasticsearchBackendDefinition{
		client: client,
	}, nil
}

func (backendDefinition *elasticsearchBackendDefinition) QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error) {
	return nil, nil
}
