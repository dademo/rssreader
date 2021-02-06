package log

import (
	"net/http"

	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/web"
)

func getLogs(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		WithFeedItems bool `httpParameter:"withFeedItems" httpParameterDefaultValue:"false"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(err, "An error occured when fetching parsing values")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	feeds, err := dbfeed.GetAllFeeds(requestParameters.WithFeedItems)
	if err != nil {
		appLog.DebugError(err, "An error occured when fetching values")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}
	web.MarshallWriteJson(responseWriter, feeds)
}
