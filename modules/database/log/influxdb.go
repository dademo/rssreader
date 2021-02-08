package log

import (
	"errors"

	"github.com/dademo/rssreader/modules/log/hook"
	client "github.com/influxdata/influxdb1-client/v2"
)

type influxDBBackendDefinition struct {
	client *client.Client
}

func influxDBLogBackend() (LogBackend, error) {

	client := hook.GetInfluxDBClient()
	if client == nil {
		return nil, errors.New("Unable to get InfluxDB client, got nil value")
	}

	return &influxDBBackendDefinition{
		client: client,
	}, nil
}

func (backendDefinition *influxDBBackendDefinition) QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error) {
	return nil, nil
}
