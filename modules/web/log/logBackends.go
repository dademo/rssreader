package log

import (
	"net/http"

	appLog "github.com/dademo/rssreader/modules/log"
	"github.com/dademo/rssreader/modules/log/hook"
	"github.com/dademo/rssreader/modules/web"
)

func getLogBackends(responseWriter http.ResponseWriter, request *http.Request) {

	var requestParameters struct {
		Backend string `httpParameter:"backend" httpParameterDefaultValue:"logrus"`
	}

	web.DisableClientCache(responseWriter)

	if err := web.ParseArgs(&requestParameters, request); err != nil {
		appLog.DebugError(err, "An error occured when fetching parsing values")
		web.AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	web.MarshallWriteJson(responseWriter, hook.GetEnabledBackends())
}
