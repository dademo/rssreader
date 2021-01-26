package web

import (
	"fmt"
	"net/http"

	appDatabase "github.com/dademo/rssreader/modules/database"
	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"
)

func init() {
	RegisterRoutes(
		RegisteredRoute{pattern: "/api/feeds", handler: getFeeds},
		RegisteredRoute{pattern: "/api/feedItems", handler: getFeedItems},
	)
}

func getFeeds(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		WithFeedItems bool `httpParameter:"withFeedItems"`
	}
	requestParameters.WithFeedItems = false

	if err := ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	feeds, err := dbfeed.GetAllFeeds(requestParameters.WithFeedItems)
	if err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	MarshallWriteJson(responseWriter, feeds)
}

func getFeedItems(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		FeedId appDatabase.PrimaryKey `httpParameter:"feed"`
	}
	requestParameters.FeedId = 0

	if err := ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching parsing values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if requestParameters.FeedId != 0 {
		feeds, err := dbfeed.GetFeedItems(requestParameters.FeedId)
		if err != nil {
			appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		MarshallWriteJson(responseWriter, feeds)
	} else {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
}
