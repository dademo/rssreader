package feeds

import "github.com/dademo/rssreader/modules/web"

func init() {
	web.RegisterRoutes(
		web.RegisteredRoute{Pattern: "/api/feed", Handler: getFeeds},
		web.RegisteredRoute{Pattern: "/api/feed/filter", Handler: filterFeeds},
		web.RegisteredRoute{Pattern: "/api/feed/{feedId}/items", Handler: getFeedItems},
		web.RegisteredRoute{Pattern: "/api/feed/{feedId}/items/filter", Handler: filterFeedItems},
	)
}
