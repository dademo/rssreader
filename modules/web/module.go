package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
)

type RegisteredRoute struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

var (
	registeredRoutes    []RegisteredRoute
	jsonContentTypeUtf8 = "application/json; charset=utf-8"
)

func RegisterServerHandlers(serveMux *http.ServeMux, httpConfig config.HttpConfig) error {

	var fileServerDir string

	if !filepath.IsAbs(httpConfig.StaticFilesDir) {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			appLog.DebugError("Unable to get application running directory")
			return err
		}
		fileServerDir = filepath.Join(appDir, httpConfig.StaticFilesDir)
	} else {
		fileServerDir = httpConfig.StaticFilesDir
	}

	fileInfo, err := os.Stat(fileServerDir)
	if err != nil {
		if os.IsNotExist(err) {
			appLog.DebugError(fmt.Sprintf("Directory specified with fileServerDir does not exists [%s]", fileServerDir))
			return err
		} else if os.IsPermission(err) {
			appLog.DebugError(fmt.Sprintf("You do not have permission to read directory [%s]", fileServerDir))
			return err
		} else {
			appLog.DebugError(fmt.Sprintf("Unknown error while stat static file serve on path [%s]", fileServerDir))
			return err
		}
	}
	if !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("You must provide a directory as the static files directory [%s]", fileServerDir))
	}

	for _, registeredRoute := range registeredRoutes {
		serveMux.HandleFunc(registeredRoute.pattern, registeredRoute.handler)
	}
	serveMux.Handle("/", http.FileServer(dotFileHidingFileSystem{http.Dir(fileServerDir)}))

	return nil
}

func MarshallWriteJson(responseWriter http.ResponseWriter, value interface{}) {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		appLog.DebugError("Unable to handle the client request")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("Content-Type", jsonContentTypeUtf8)

	_, err = responseWriter.Write(marshalledValue)
	if err != nil {
		appLog.DebugError("Unable to write answer")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RegisterRoutes(routes ...RegisteredRoute) {
	registeredRoutes = append(registeredRoutes, routes...)
}
