package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dademo/rssreader/modules/log"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type LogHandler struct {
	wrappedHandler http.Handler
}

func (handler LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var uuidStr string
	var strStatusCode string
	uuid, err := uuid.NewRandom()

	if err == nil {
		uuidStr = uuid.String()
	} else {
		log.DebugError(err, "Unable to create uuid")
		uuidStr = "-"
	}

	queried := r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		queried += "?" + r.URL.RawQuery
	}
	if len(r.URL.RawFragment) > 0 {
		queried += "#" + r.URL.RawFragment
	}

	logrus.WithFields(logrus.Fields{
		"uuid":      uuidStr,
		"method":    r.Method,
		"protocol":  r.Proto,
		"scheme":    r.URL.Scheme,
		"host":      r.Host,
		"url":       r.URL.Path,
		"query":     r.URL.RawQuery,
		"fragment":  r.URL.RawFragment,
		"userAgent": r.UserAgent(),
	}).Trace("Received request")

	startedAt := time.Now()
	handler.wrappedHandler.ServeHTTP(w, r)

	if r.Response != nil {
		strStatusCode = strconv.Itoa(r.Response.StatusCode)
	} else {
		strStatusCode = ""
	}

	logrus.Info(
		fmt.Sprintf("[%s]\t%s %s %d",
			r.Method,
			queried,
			strStatusCode,
			r.ContentLength,
		),
	)

	endedAt := time.Now()
	logrus.WithFields(logrus.Fields{
		"uuid":                uuidStr,
		"method":              r.Method,
		"protocol":            r.Proto,
		"scheme":              r.URL.Scheme,
		"host":                r.URL.Host,
		"url":                 r.URL.Path,
		"query":               r.URL.RawQuery,
		"fragment":            r.URL.RawFragment,
		"statusCode":          strStatusCode,
		"contentLength":       r.ContentLength,
		"durationNanoSeconds": endedAt.UnixNano() - startedAt.UnixNano(),
	}).Trace("Request processed")
}

func HttpLogInterceptorFor(handler http.Handler) http.Handler {
	return LogHandler{
		wrappedHandler: handler,
	}
}
