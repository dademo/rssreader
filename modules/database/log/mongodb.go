package log

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/log/hook"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const DEFAULT_QUERY_TIMEOUT = 5 * time.Second

//// MongoDB Comparators ////
// https://docs.mongodb.com/manual/reference/operator/query-comparison/
const (
	mongoComparatorGreaterEqualThan = "$gte"
	mongoComparatorGreaterThan      = "$gt"
	mongoComparatorEquals           = "$eq"
	mongoComparatorIn               = "$in"
	mongoComparatorLowerThan        = "$lt"
	mongoComparatorLowerEqualThan   = "$lte"
	mongoComparatorNonEqual         = "$ne"
	mongoComparatorNonIn            = "$nin"
)

type mongoDBBackendDefinition struct {
	client *mongo.Client
	config *config.MongoDBLogBackendConfig
}

type LogEntryMongoDB struct {
	Id        string
	Timestamp int64
	Level     string
	File      string
	Function  string
	Message   string
	Data      LogDataEnties
}

func mongoDBLogBackend() (LogBackend, error) {

	client := hook.GetMongoDBClient()
	if client == nil {
		return nil, errors.New("Unable to get MongoDB client, got nil value")
	}

	return &mongoDBBackendDefinition{
		client: client,
		config: hook.GetMongoDBConfig(),
	}, nil
}

func (backendDefinition *mongoDBBackendDefinition) QueryForLogs(query *LogQueryOpts) (*LogEntriesPage, error) {

	var logEntriesPage *LogEntriesPage
	var totalElements int64
	var err error
	backendTimeout := time.Duration(backendDefinition.config.TimeoutSeconds) * time.Second

	mongoQuery, err := makeMongoDBQuery(query)
	if err != nil {
		return nil, fmt.Errorf("Unable to create query, %s", err)
	}

	ctxPing, cancelPing := context.WithTimeout(
		context.Background(),
		backendTimeout,
	)
	defer cancelPing()

	err = backendDefinition.client.Ping(ctxPing, readpref.Primary())
	if err != nil {
		log.LoggerFallback().Error("An error occured when pinging the MongoDB backend")
		return nil, err
	}

	logDatabase := backendDefinition.client.Database(backendDefinition.config.Database)
	logCollection := logDatabase.Collection(backendDefinition.config.Collection)

	logEntriesPage, err = executeQuery(backendDefinition, query, logCollection, mongoQuery)
	if err != nil {
		return nil, err
	}
	totalElements, err = getElementsCount(backendDefinition, query, logCollection, mongoQuery)
	if err != nil {
		return nil, err
	}
	logEntriesPage.TotalElements = uint(totalElements)

	return logEntriesPage, nil
}

func executeQuery(backendDefinition *mongoDBBackendDefinition,
	query *LogQueryOpts, logCollection *mongo.Collection, mongoQuery *bson.M) (*LogEntriesPage, error) {

	var err error
	logEntriesPage := LogEntriesPage{
		PageNo:        query.Page.PageNo,
		PageSize:      query.Page.PageSize,
		TotalElements: 0,
		Entries:       make([]*LogEntry, 0),
	}
	backendTimeout := time.Duration(backendDefinition.config.TimeoutSeconds) * time.Second
	queryOpts := options.Find()
	queryOpts.SetSkip(int64(query.Page.PageNo * query.Page.PageSize))
	queryOpts.SetLimit(int64(query.Page.PageSize))
	queryOpts.SetSort(bson.M{"_id": -1})

	ctxQuery, cancelCtxQuery := context.WithTimeout(
		context.Background(),
		backendTimeout,
	)
	defer cancelCtxQuery()

	cur, err := logCollection.Find(ctxQuery, *mongoQuery, queryOpts)
	if err != nil {
		return nil, fmt.Errorf("Unable while performing query, %s", err)
	}
	defer cur.Close(ctxQuery)

	for cur.Next(ctxQuery) {
		var resultMongoDB LogEntryMongoDB
		err := cur.Decode(&resultMongoDB)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode a row result, %s", err)
		}
		logEntriesPage.Entries = append(logEntriesPage.Entries, mapMongoDBResult(&resultMongoDB))
	}

	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("Unable while performing query, %s", err)
	}

	return &logEntriesPage, nil
}

func getElementsCount(backendDefinition *mongoDBBackendDefinition,
	query *LogQueryOpts, logCollection *mongo.Collection, mongoQuery *bson.M) (int64, error) {

	backendTimeout := time.Duration(backendDefinition.config.TimeoutSeconds) * time.Second

	ctxQuery, cancelCtxQuery := context.WithTimeout(
		context.Background(),
		backendTimeout,
	)
	defer cancelCtxQuery()

	return logCollection.CountDocuments(ctxQuery, *mongoQuery)
}

func makeMongoDBQuery(query *LogQueryOpts) (*bson.M, error) {

	logrusQueryLevel, err := logrus.ParseLevel(query.Level)
	if err != nil {
		return nil, err
	}

	queryFilters := []bson.M{
		{"level": bson.M{"$in": expectedLogLevels(logrusQueryLevel, query.LevelComparator)}},
		mongoDBTextCompare("message", query.Message, query.MessageComparator),
		mongoDBTextCompare("file", query.File, query.FileComparator),
		mongoDBTextCompare("func", query.Function, query.FunctionComparator),
	}

	if !query.Date.IsZero() {
		document := bson.M{}
		document[mongoDBComparatorOf(query.DateComparator)] = query.Date.UnixNano()
		queryFilters = append(queryFilters, bson.M{"timestamp": document})
	}

	if query.MatchingDataKeys != nil {
		for k, v := range query.MatchingDataKeys {
			document := bson.M{}
			document[k] = v
			queryFilters = append(queryFilters, document)
		}
	}

	return &bson.M{"$and": queryFilters}, nil
}

func mongoDBTextCompare(field string, value string, comparator database.StringComparator) bson.M {

	document := bson.M{}
	switch comparator {
	default:
		fallthrough
	case database.StrComparatorContains:
		document[field] = mongoDBTextContainsQuery(value)
	case database.StrComparatorMatches:
		document[field] = mongoDBRegexpQuery(value)
	}
	return document
}

func mongoDBTextContainsQuery(value string) bson.M {
	return bson.M{"$regex": fmt.Sprintf(".*%s.*", regexp.QuoteMeta(value))}
}

func mongoDBRegexpQuery(value string) bson.M {
	return bson.M{"$regex": value}
}

func mongoDBComparatorOf(comparator database.Comparator) string {

	switch comparator {
	default:
		fallthrough
	case database.ComparatorGreaterEqualThan:
		return mongoComparatorGreaterEqualThan
	case database.ComparatorGreaterThan:
		return mongoComparatorGreaterThan
	case database.ComparatorEquals:
		return mongoComparatorEquals
	case database.ComparatorLowerThan:
		return mongoComparatorLowerThan
	case database.ComparatorLowerEqualThan:
		return mongoComparatorLowerEqualThan
	}
}

func expectedLogLevels(logLevel logrus.Level, comparator database.Comparator) []string {

	levelStrs := make([]string, 0)

	for _, logrusLevel := range logrus.AllLevels {
		if compare(int64(logLevel)-int64(logrusLevel), comparator) {
			levelStrs = append(levelStrs, strings.ToUpper(logrusLevel.String()))
		}
	}

	return levelStrs
}

func mapMongoDBResult(mongoDBResult *LogEntryMongoDB) *LogEntry {

	logEntry := LogEntry{}

	logEntry.Level = mongoDBResult.Level
	logEntry.Timestamp = time.Unix(0, mongoDBResult.Timestamp)
	logEntry.Data = mongoDBResult.Data
	logEntry.Message = mongoDBResult.Message
	logEntry.File = mongoDBResult.File
	logEntry.Function = mongoDBResult.Function

	return &logEntry
}
