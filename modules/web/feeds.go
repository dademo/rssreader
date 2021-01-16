package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dademo/rssreader/modules/database/dbfeed"
	appLog "github.com/dademo/rssreader/modules/log"

	log "github.com/sirupsen/logrus"
)

func init() {
	RegisterRoutes(
		RegisteredRoute{pattern: "/api/feeds", handler: getFeeds},
	)
}

func getFeeds(responseWriter http.ResponseWriter, request *http.Request) {

	var withFeedItems bool
	var err error

	withFeedItemsStr := request.URL.Query().Get("withFeedItems")

	if withFeedItemsStr != "" {
		withFeedItems, err = strconv.ParseBool(withFeedItemsStr)
		if err != nil {
			log.Warning(fmt.Sprintf("Unparseable parameter for withFeedItems parameter (got %s)", withFeedItemsStr))
			withFeedItems = false
		}
	} else {
		withFeedItems = false
	}

	feeds, err := dbfeed.GetAllFeeds(withFeedItems)
	if err != nil {
		appLog.DebugError(fmt.Sprintf("An error occured when fetching values (%s)", err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	MarshallWriteJson(responseWriter, feeds)
}
