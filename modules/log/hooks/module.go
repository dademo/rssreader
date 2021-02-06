package hooks

import (
	"time"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v7"

	// Pre-built hooks
	"github.com/Abramovic/logrus_influxdb"
)

func RegisterHooks(logBackendsConfig *config.LogBackendsDefinition) (err error) {

	logrus.Debug("Registering log hooks")

	if logBackendsConfig.Elasticsearch.Enabled {
		err = RegisterElasticsearchHook(logBackendsConfig.Elasticsearch)
		if err != nil {
			appLog.LoggerFallback().WithError(err).Error("Unable to register Elasticsearch backend")
			return err
		}
	}

	if logBackendsConfig.InfluxDB.Enabled {
		err = RegisterInfluxDBHook(logBackendsConfig.InfluxDB)
		if err != nil {
			appLog.LoggerFallback().WithError(err).Error("Unable to register InfluxDB backend")
			return err
		}
	}

	if logBackendsConfig.MongoDB.Enabled {
		err = RegisterMongoDBHook(logBackendsConfig.MongoDB)
		if err != nil {
			appLog.LoggerFallback().WithError(err).Error("Unable to register MongoDB backend")
			return err
		}
	}

	if logBackendsConfig.BuntDB.Enabled {
		err = RegisterBuntDBHook(logBackendsConfig.BuntDB)
		if err != nil {
			appLog.LoggerFallback().WithError(err).Error("Unable to register BuntDB backend")
			return err
		}
	}

	logrus.Debug("Log hooks registered")

	return nil
}

func RegisterElasticsearchHook(config *config.ElasticsearchLogBackendConfig) error {

	client, err := connectElasticsearchClient(config)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to create logrus Elasticsearch hook connection")
		return err
	}

	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, config.Index)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to create logrus Elasticsearch hook")
		return err
	}

	logrus.AddHook(hook)
	return nil
}

func RegisterInfluxDBHook(config *config.InfluxDBLogBackendConfig) error {

	client, err := connectInfluxDBClient(config)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to create logrus InfluxDB hook connection")
		return err
	}

	hookConfig := &logrus_influxdb.Config{
		Host:          config.Host,
		Timeout:       time.Duration(config.TimeoutSeconds) * time.Second,
		Database:      config.Database,
		Username:      config.Username,
		Password:      config.Password,
		Precision:     config.Precision,
		MinLevel:      config.LogMinLevel,
		Measurement:   config.Measurement,
		BatchInterval: (time.Duration(config.BatchIntervalSeconds) * time.Second),
		BatchCount:    config.BatchCount,
	}

	hook, err := logrus_influxdb.NewInfluxDB(hookConfig, *client)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to create logrus Influxdb hook")
		return err
	}

	logrus.AddHook(hook)
	return nil
}

func RegisterMongoDBHook(config *config.MongoDBLogBackendConfig) error {

	logLevels, err := getLevelsGreaterThan(config.LogMinLevel)
	if err != nil {
		return err
	}

	mongoClient, err := connectMongo(config)
	if err != nil {
		return err
	}

	logDatabase := mongoClient.Database(config.Database)
	logCollection := logDatabase.Collection(config.Collection)

	logrus.AddHook(MongoDBLogHook{
		config:        config,
		logCollection: logCollection,
		levels:        logLevels,
	})

	return nil
}

func RegisterBuntDBHook(config *config.BuntDBLogBackendConfig) error {

	logLevels, err := getLevelsGreaterThan(config.LogMinLevel)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to create logrus BuntDB hook connection")
		return err
	}

	client, err := connectBuntDB(config)
	if err != nil {
		return err
	}

	logrus.AddHook(BuntDBLogHook{
		config: config,
		db:     client,
		levels: logLevels,
	})
	return nil
}

func getLevelsGreaterThan(levelStr string) ([]logrus.Level, error) {

	minLevel, err := logrus.ParseLevel(levelStr)
	if err != nil {
		appLog.LoggerFallback().WithError(err).Error("Unable to parse log level")
		return []logrus.Level{}, err
	}

	levels := make([]logrus.Level, 0)
	for _, level := range logrus.AllLevels {
		if level <= minLevel {
			levels = append(levels, level)
		}
	}

	return levels, nil
}

func mergeAdditionalTags(tags ...map[string]interface{}) map[string]interface{} {

	finalTags := map[string]interface{}{}
	for _, tag := range tags {
		for k, v := range tag {
			finalTags[k] = v
		}
	}
	return finalTags
}
