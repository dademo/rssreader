package feeds

import (
	"fmt"
	"net/http"

	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/web"
)

func getFeeds(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		WithFeedItems bool `httpParameter:"withFeedItems" httpParameterDefaultValue:"false"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	feeds, err := dbfeed.GetAllFeeds(requestParameters.WithFeedItems)
	if err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}
	web.MarshallWriteJson(responseWriter, feeds)
}

func filterFeeds(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		WithFeedItems bool   `httpParameter:"withFeedItems" httpParameterDefaultValue:"false"`
		Field         string `httpParameter:"field" httpParameterDefaultValue:""`
		Filter        string `httpParameter:"filter"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	feeds, err := dbfeed.GetAllFeeds(requestParameters.WithFeedItems)
	if err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}
	web.MarshallWriteJson(responseWriter, feeds)
}
