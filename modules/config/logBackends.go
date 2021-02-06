package config

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type LogBackendsDefinition struct {
	Elasticsearch *ElasticsearchLogBackendConfig `yaml:"elasticsearch"`
	InfluxDB      *InfluxDBLogBackendConfig      `yaml:"influxDB"`
	MongoDB       *MongoDBLogBackendConfig       `yaml:"mongoDB"`
	BuntDB        *BuntDBLogBackendConfig        `yaml:"buntDB"`
}

type ElasticsearchLogBackendConfig struct {
	Enabled     bool     `yaml:"enabled"`
	LogMinLevel string   `yaml:"logMinLevel"`
	Addresses   []string `yaml:"addresses"`
	Host        string   `yaml:"host"`
	Index       string   `yaml:"index"`

	// Authentication informations
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	CloudID  string `yaml:"cloudID"`
	APIKey   string `yaml:"apiKey"`

	// Elasticsearch specific configuration
	Header               http.Header `yaml:"header"`
	CACertFile           string      `yaml:"caCertFile"`
	RetryOnStatus        []int       `yaml:"retryOnStatus"`
	DisableRetry         bool        `yaml:"disableRetry"`
	EnableRetryOnTimeout bool        `yaml:"enableRetryOnTimeout"`
	MaxRetries           int         `yaml:"maxRetries"`
}

type InfluxDBLogBackendConfig struct {
	Enabled              bool   `yaml:"enabled"`
	LogMinLevel          string `yaml:"logMinLevel"`
	Url                  string `yaml:"url"`
	Host                 string `yaml:"host"`
	TimeoutSeconds       int    `yaml:"timeoutSeconds"`
	Database             string `yaml:"database"`
	Username             string `yaml:"username"`
	Password             string `yaml:"password"`
	Precision            string `yaml:"precision"`
	Measurement          string `yaml:"measurement"`
	BatchIntervalSeconds int    `yaml:"batchIntervalSeconds"`
	BatchCount           int    `yaml:"batchCount"`
}

type MongoDBLogBackendConfig struct {
	Enabled        bool                   `yaml:"enabled"`
	LogMinLevel    string                 `yaml:"logMinLevel"`
	URI            string                 `yaml:"uri"`
	Database       string                 `yaml:"database"`
	Collection     string                 `yaml:"collection"`
	TimeoutSeconds int                    `yaml:"timeoutSeconds"`
	AdditionalTags map[string]interface{} `yaml:"additionalTags"`
}

type BuntDBLogBackendConfig struct {
	Enabled           bool                   `yaml:"enabled"`
	LogMinLevel       string                 `yaml:"logMinLevel"`
	File              string                 `yaml:"file"`
	ExpirationSeconds int                    `yaml:"expirationSeconds"`
	AdditionalTags    map[string]interface{} `yaml:"additionalTags"`
}

func defaultLogBackendsConfig() *LogBackendsDefinition {
	return &LogBackendsDefinition{
		Elasticsearch: defaultElasticsearchLogBackendConfig(),
		InfluxDB:      defaultInfluxDBLogBackendConfig(),
		MongoDB:       defaultMongoDBLogBackendConfig(),
		BuntDB:        defaultBuntDBLogBackendConfig(),
	}
}

func defaultElasticsearchLogBackendConfig() *ElasticsearchLogBackendConfig {
	return &ElasticsearchLogBackendConfig{
		Enabled:              false,
		LogMinLevel:          logrus.TraceLevel.String(),
		Addresses:            []string{"http://localhost:9200"},
		Host:                 "localhost",
		Index:                "rss-reader",
		RetryOnStatus:        []int{},
		DisableRetry:         true,
		EnableRetryOnTimeout: true,
		MaxRetries:           3,
	}
}

func defaultInfluxDBLogBackendConfig() *InfluxDBLogBackendConfig {
	return &InfluxDBLogBackendConfig{
		Enabled:              false,
		LogMinLevel:          logrus.TraceLevel.String(),
		Url:                  "http://localhost:8086",
		Host:                 "localhost",
		TimeoutSeconds:       5,
		Database:             "rss_reader",
		Username:             "rssreader",
		Password:             "rssreader",
		Precision:            "ms",
		Measurement:          "logs",
		BatchIntervalSeconds: 5,
		BatchCount:           0,
	}
}

func defaultMongoDBLogBackendConfig() *MongoDBLogBackendConfig {
	return &MongoDBLogBackendConfig{
		Enabled:        false,
		LogMinLevel:    logrus.TraceLevel.String(),
		URI:            "mongodb:///rssreader:rssreader@localhost:27017",
		Database:       "rssreader",
		Collection:     "logs",
		TimeoutSeconds: 5,
		AdditionalTags: map[string]interface{}{},
	}
}

func defaultBuntDBLogBackendConfig() *BuntDBLogBackendConfig {
	return &BuntDBLogBackendConfig{
		Enabled:           false,
		LogMinLevel:       logrus.TraceLevel.String(),
		File:              ":memory:",
		ExpirationSeconds: 0,
		AdditionalTags:    map[string]interface{}{},
	}
}
