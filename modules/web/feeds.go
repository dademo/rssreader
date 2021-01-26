package web

import (
	"fmt"
	"net/http"

	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"
)

func init() {
	RegisterRoutes(
		RegisteredRoute{pattern: "/api/feeds", handler: getFeeds},
	)
}

func getFeeds(responseWriter http.ResponseWriter, request *http.Request) {

	var requestBody struct {
		WithFeedItems bool `httpParameter:"withFeedItems"`
	}

	if err := ParseArgs(&requestBody, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	feeds, err := dbfeed.GetAllFeeds(requestBody.WithFeedItems)
	if err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	MarshallWriteJson(responseWriter, feeds)
}
