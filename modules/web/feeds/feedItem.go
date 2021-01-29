package feeds

import (
	"fmt"
	"net/http"

	appDatabase "github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/web"
)

func getFeedItems(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		FeedId appDatabase.PrimaryKey `httpParameter:"feedId"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	if requestParameters.FeedId != 0 {
		feeds, err := dbfeed.GetFeedItems(requestParameters.FeedId)
		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
			web.AnswerError(err, http.StatusInternalServerError, responseWriter)
			return
		}
		web.MarshallWriteJson(responseWriter, feeds)
	} else {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
}

func filterFeedItems(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		FeedId appDatabase.PrimaryKey `httpParameter:"feedId"`
		Field  string                 `httpParameter:"field" httpParameterDefaultValue:""`
		Filter string                 `httpParameter:"filter"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	if requestParameters.FeedId != 0 {
		feeds, err := dbfeed.GetFeedItems(requestParameters.FeedId)
		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
			web.AnswerError(err, http.StatusInternalServerError, responseWriter)
			return
		}
		web.MarshallWriteJson(responseWriter, feeds)
	} else {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
}
