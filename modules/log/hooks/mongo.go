package hooks

import (
	"context"
	"strings"
	"time"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBLogHook struct {
	config        *config.MongoDBLogBackendConfig
	logCollection *mongo.Collection
	levels        []logrus.Level
}

type MongoDBLogEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp int64              `bson:"timestamp,omitempty"`
	Level     string             `bson:"level,omitempty"`
	File      string             `bson:"file,omitempty"`
	Function  string             `bson:"func,omitempty"`
	Message   string             `bson:"message,omitempty"`
	Data      logrus.Fields      `bson:"data,omitempty"`
}

var mongoDBClient *mongo.Client

func connectMongo(config *config.MongoDBLogBackendConfig) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		mongoOptions.Client().ApplyURI(config.URI),
	)
	if err != nil {
		appLog.LoggerFallback().Error("Unable to connect to the MongoDB database")
		return nil, err
	}

	ctxPing, cancelPing := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancelPing()
	err = client.Ping(ctxPing, readpref.Primary())
	if err != nil {
		appLog.LoggerFallback().Error("An error occured when pinging the MongoDB backend")
		return nil, err
	}

	logrus.RegisterExitHandler(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
		defer cancelCleanup()
		closeErr := client.Disconnect(cleanupCtx)
		if closeErr != nil {
			appLog.LoggerFallback().WithError(closeErr).Error("An error occured when closing MongoDB database")
		}
	})

	mongoDBClient = client

	return client, nil
}

func (hook MongoDBLogHook) Fire(entry *logrus.Entry) error {

	var file string
	var function string
	if entry.HasCaller() {
		file = entry.Caller.File
		function = entry.Caller.Function
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(hook.config.TimeoutSeconds)*time.Second,
	)
	defer cancel()

	mongoEntry := MongoDBLogEntry{
		Timestamp: entry.Time.UnixNano(),
		Level:     strings.ToUpper(entry.Level.String()),
		File:      file,
		Function:  function,
		Message:   entry.Message,
		Data:      mergeAdditionalTags(hook.config.AdditionalTags, entry.Data),
	}

	_, err := hook.logCollection.InsertOne(ctx, mongoEntry)
	if err != nil {
		appLog.LoggerFallback().Error("Unable to add log entry to MongoDB collection")
		return err
	}

	return nil
}

func (hook MongoDBLogHook) Levels() []logrus.Level {
	return hook.levels
}

func GetMongoDBClient() *mongo.Client {
	return mongoDBClient
}
