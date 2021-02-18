package log

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dademo/rssreader/modules/database"
	dbLog "github.com/dademo/rssreader/modules/database/log"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/web"
	log "github.com/sirupsen/logrus"
)

type getLogsParameters struct {
	Backend string `httpParameter:"backend" httpParameterDefaultValue:"buntdb"`

	PageNo   uint   `httpParameter:"pageNo" httpParameterDefaultValue:"0"`
	PageSize uint   `httpParameter:"pageSize" httpParameterDefaultValue:"0"`
	Sort     string `httpParameter:"sort" httpParameterDefaultValue:""`

	Level           string `httpParameter:"level" httpParameterDefaultValue:"info"`
	LevelComparator string `httpParameter:"levelCompare" httpParameterDefaultValue:"GE"`

	Date           string `httpParameter:"date" httpParameterDefaultValue:""`
	DateComparator string `httpParameter:"dateCompare" httpParameterDefaultValue:"GE"`

	Message           string `httpParameter:"message" httpParameterDefaultValue:""`
	MessageComparator string `httpParameter:"messageCompare" httpParameterDefaultValue:"CONTAINS"`

	File           string `httpParameter:"file" httpParameterDefaultValue:""`
	FileComparator string `httpParameter:"fileCompare" httpParameterDefaultValue:"CONTAINS"`

	Function           string `httpParameter:"function" httpParameterDefaultValue:""`
	FunctionComparator string `httpParameter:"functionCompare" httpParameterDefaultValue:"CONTAINS"`

	MatchingDataKeys string `httpParameter:"matchingDataKeys" httpParameterDefaultValue:""`

	ScrollID string `httpParameter:"scrollId" httpParameterDefaultValue:""`
}

const DataKeysSeparator = ","

var TimeFormats = []string{
	"2006-01-02T15:04:05",      // ISO 8601 / RFC3339
	"2006-01-02T15:04:05-0600", // ISO 8601 / RFC3339
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02 03:04:05",
	"2006-01-02",
}

func getLogs(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters getLogsParameters

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(err, "An error occured when fetching parsing values")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	backend, err := dbLog.GetLogBackendFromString(requestParameters.Backend)
	if err != nil {
		appLog.DebugError(err, "An error occured when getting backend")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	query, err := parseDocumentToQuery(&requestParameters)
	if err != nil {
		appLog.DebugError(err, "An error occured when parsing the query")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	logPage, err := backend.QueryForLogs(query)
	if err != nil {
		appLog.DebugError(err, "An error occured when querying for values")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}
	web.MarshallWriteJson(responseWriter, logPage)
}

func parseDocumentToQuery(requestParameters *getLogsParameters) (*dbLog.LogQueryOpts, error) {

	page, err := parsePage(requestParameters)
	if err != nil {
		return nil, err
	}

	logLevelComparator, err := database.ParseComparator(requestParameters.LevelComparator)
	if err != nil {
		return nil, err
	}

	logDate, err := parseLogDate(requestParameters.Date)
	if err != nil {
		return nil, err
	}

	logDateComparator, err := database.ParseComparator(requestParameters.DateComparator)
	if err != nil {
		return nil, err
	}

	messageComparator, err := database.ParseStringComparator(requestParameters.MessageComparator)
	if err != nil {
		return nil, err
	}

	fileComparator, err := database.ParseStringComparator(requestParameters.FileComparator)
	if err != nil {
		return nil, err
	}

	functionComparator, err := database.ParseStringComparator(requestParameters.FunctionComparator)
	if err != nil {
		return nil, err
	}

	matchingDataKeys, err := parseMatchingDataKeys(requestParameters.MatchingDataKeys)
	if err != nil {
		return nil, err
	}

	return &dbLog.LogQueryOpts{
		Page: page,

		Level: strings.ToUpper(fistNonEmptyStr(
			requestParameters.Level,
			log.InfoLevel.String(),
		)),
		LevelComparator: logLevelComparator,

		Date:           logDate,
		DateComparator: logDateComparator,

		Message:           requestParameters.Message,
		MessageComparator: messageComparator,

		File:           requestParameters.File,
		FileComparator: fileComparator,

		Function:           requestParameters.Function,
		FunctionComparator: functionComparator,

		MatchingDataKeys: matchingDataKeys,
		ScrollID:         requestParameters.ScrollID,
	}, nil
}

func parsePage(requestParameters *getLogsParameters) (database.PageQuery, error) {

	sort, err := database.ParseSortOptions(requestParameters.Sort)
	if err != nil {
		log.Error("Unable to parse page sort options")
		return database.PageQuery{}, err
	}

	return database.PageQuery{
		PageNo:   requestParameters.PageNo,
		PageSize: requestParameters.PageSize,
		Sort:     sort,
	}, nil
}

func parseLogDate(logDateStr string) (time.Time, error) {

	if logDateStr == "" {
		return time.Time{}, nil
	}

	for _, timeFormat := range TimeFormats {
		parsedTime, err := time.Parse(timeFormat, logDateStr)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"Unable to parse date [%s]. Available formats are '%s'",
		logDateStr, strings.Join(TimeFormats, ", "),
	)
}

func parseMatchingDataKeys(matchingDataKeysStr string) (map[string]interface{}, error) {

	result := map[string]interface{}{}
	// Format : "key1,value1,key2,value2"
	splittedKeyValues := strings.Split(matchingDataKeysStr, DataKeysSeparator)

	if len(splittedKeyValues)%2 != 0 {
		if splittedKeyValues[len(splittedKeyValues)-1] != "" {
			return result, errors.New("Bad number of key-values for data keys order")
		} else {
			splittedKeyValues = splittedKeyValues[:len(splittedKeyValues)-1]
		}
	}

	for it := 0; it < len(splittedKeyValues); it += 2 {
		if splittedKeyValues[it] != "" {
			result[splittedKeyValues[it]] = splittedKeyValues[it+1]
		} else if it != len(splittedKeyValues)-1 { // If not the last element, raising error; we skip if last because last key can be an empty string
			return nil, errors.New("No key given for a matching data-key value key")
		}
	}

	return result, nil
}

func fistNonEmptyStr(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
