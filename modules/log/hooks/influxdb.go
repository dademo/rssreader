package hooks

import (
	"time"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/sirupsen/logrus"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var influxdbClient *influxdb.Client

func connectInfluxDBClient(config *config.InfluxDBLogBackendConfig) (*influxdb.Client, error) {

	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     config.Url,
		Username: config.Username,
		Password: config.Password,
		Timeout:  time.Duration(config.TimeoutSeconds) * time.Second,
	})
	if err != nil {
		appLog.DebugError(err, "Unable to establish a InfluxDB connection")
		return nil, err
	}

	_, _, err = client.Ping(time.Duration(config.TimeoutSeconds) * time.Second)
	if err != nil {
		appLog.DebugError(err, "An error occured when pinging the InfluxDB database")
		return nil, err
	}

	logrus.RegisterExitHandler(func() {
		closeErr := client.Close()
		if closeErr != nil {
			appLog.LoggerFallback().WithError(closeErr).Error("An error occured when closing InfluxDB database")
		}
	})

	influxdbClient = &client

	return &client, nil
}

func GetInfluxDBClient() *influxdb.Client {
	return influxdbClient
}
