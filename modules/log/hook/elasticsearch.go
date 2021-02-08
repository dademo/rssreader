package hook

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/elastic/go-elasticsearch/v7"
)

var elasticsearchClient *elasticsearch.Client

func connectElasticsearchClient(config *config.ElasticsearchLogBackendConfig) (*elasticsearch.Client, error) {

	if config == nil {
		return nil, errors.New("No configuration given for elasticsearch logrus hook")
	}

	var caCertContent []byte
	var err error
	if config.CACertFile != "" {
		if caCertContent, err = ioutil.ReadFile(config.CACertFile); err != nil {
			appLog.LoggerFallback().WithError(err).Error(fmt.Sprintf("Unable to readcertificate at path [%s]", config.CACertFile))
			return nil, err

		}
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:            config.Addresses,
		Username:             config.Username,
		Password:             config.Password,
		CloudID:              config.CloudID,
		Header:               config.Header,
		CACert:               caCertContent,
		RetryOnStatus:        config.RetryOnStatus,
		DisableRetry:         config.DisableRetry,
		EnableRetryOnTimeout: config.EnableRetryOnTimeout,
		MaxRetries:           config.MaxRetries,
	})
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to connect to the Elasticsearch service")
		return nil, err
	}

	elasticsearchClient = client

	return client, nil
}

func GetElasticsearchClient() *elasticsearch.Client {
	return elasticsearchClient
}
