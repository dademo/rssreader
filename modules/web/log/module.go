package log

import (
	"github.com/dademo/rssreader/modules/web"
)

func init() {
	web.RegisterRoutes(
		web.RegisteredRoute{Pattern: "/api/log/backends", Handler: getLogBackends},
		web.RegisteredRoute{Pattern: "/api/log", Handler: getLogs},
	)
}
