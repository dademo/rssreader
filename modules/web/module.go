package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
)

type RegisteredRoute struct {
	Pattern string
	Handler func(http.ResponseWriter, *http.Request)
}

var (
	registeredRoutes []RegisteredRoute
	displayErrors    bool = false
)

const (
	AppApiPrefix                     = "/api"
	HTTPParameterTagName             = "httpParameter"
	HTTPParameterDefaultValueTagName = "httpParameterDefaultValue"
	JSONContentTypeUtf8              = "application/json; charset=utf-8"
	TextPlainUTF8                    = "text/plain; charset=utf-8"
	headerContentType                = "Content-Type"
	contentTypeMultipartFormData     = "multipart/form-data"
)

func RegisterServerHandlers(serveMux *http.ServeMux, httpConfig *config.HttpConfig) error {

	var fileServerDir string

	if !filepath.IsAbs(httpConfig.StaticFilesDir) {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			appLog.DebugError(err, "Unable to get application running directory")
			return err
		}
		fileServerDir = filepath.Join(appDir, httpConfig.StaticFilesDir)
	} else {
		fileServerDir = httpConfig.StaticFilesDir
	}

	fileInfo, err := os.Stat(fileServerDir)
	if err != nil {
		if os.IsNotExist(err) {
			appLog.DebugError(err, fmt.Sprintf("Directory specified with fileServerDir does not exists [%s]", fileServerDir))
			return err
		} else if os.IsPermission(err) {
			appLog.DebugError(err, fmt.Sprintf("You do not have permission to read directory [%s]", fileServerDir))
			return err
		} else {
			appLog.DebugError(err, fmt.Sprintf("Unknown error while stat static file serve on path [%s]", fileServerDir))
			return err
		}
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("You must provide a directory as the static files directory [%s]", fileServerDir)
	}

	router := mux.NewRouter()
	for _, registeredRoute := range registeredRoutes {
		router.HandleFunc(registeredRoute.Pattern, registeredRoute.Handler)
	}

	router.PathPrefix("/").Handler(http.FileServer(dotFileHidingFileSystem{http.Dir(fileServerDir)}))
	serveMux.Handle("/", HttpLogInterceptorFor(router))

	return nil
}

func MarshallWriteJson(responseWriter http.ResponseWriter, value interface{}) {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		appLog.DebugError(err, "Unable to handle the client request")
		AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}

	responseWriter.Header().Add("Content-Type", JSONContentTypeUtf8)

	_, err = responseWriter.Write(marshalledValue)
	if err != nil {
		appLog.DebugError(err, "Unable to write answer")
		AnswerError(err, http.StatusInternalServerError, responseWriter)
		return
	}
}

func RegisterRoutes(routes ...RegisteredRoute) {
	registeredRoutes = append(registeredRoutes, routes...)
}

func ParseArgs(arguments interface{}, request *http.Request) error {

	var urlParameter string

	if headerContentTypeValue := request.Header.Get(headerContentType); headerContentTypeValue == contentTypeMultipartFormData {
		err := request.ParseMultipartForm(0)
		if err != nil {
			return err
		}
	} else {
		err := request.ParseForm()
		if err != nil {
			return err
		}
	}

	requestParameters := mux.Vars(request)
	/* Mering maps */
	for k, _ := range request.Form {
		requestParameters[k] = request.Form.Get(k)
	}

	reflected := reflect.ValueOf(arguments)

	if reflected.IsNil() {
		return errors.New("Provided value should not be nil")
	}

	if reflected.Type().Kind() != reflect.Ptr {
		return errors.New("You must provide a pointer to parse arguments")
	}

	elem := reflected.Elem()

	for i := 0; i < elem.NumField(); i++ {
		typeField := elem.Type().Field(i)
		tag := typeField.Tag

		valueField := elem.Field(i)

		if !valueField.CanSet() {
			return fmt.Errorf("Unable to set value for field [%s]", typeField.Name)
		}

		if tag, ok := tag.Lookup(HTTPParameterTagName); ok {
			urlParameter = tag
		} else {
			urlParameter = typeField.Name
		}

		if strURLValue, ok := requestParameters[urlParameter]; ok && strURLValue != "" {
			err := setFieldValue(valueField, strURLValue)
			if err != nil {
				return err
			}
		} else if tag, ok := tag.Lookup(HTTPParameterDefaultValueTagName); ok {
			err := setFieldValue(valueField, tag)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Field [%s] (http parameter '%s') is missing and has no default value", typeField.Name, urlParameter)
		}
	}

	return nil
}

func setFieldValue(valueField reflect.Value, strValue string) error {

	switch valueField.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, err := strconv.ParseInt(strValue, 10, 0); err != nil {
			return err
		} else {
			valueField.SetInt(v)
		}
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, err := strconv.ParseUint(strValue, 10, 0); err != nil {
			return err
		} else {
			valueField.SetUint(v)
		}
		break
	case reflect.Float32, reflect.Float64:
		if v, err := strconv.ParseFloat(strValue, 0); err != nil {
			return err
		} else {
			valueField.SetFloat(v)
		}
		break
	case reflect.Bool:
		if v, err := strconv.ParseBool(strValue); err != nil {
			return err
		} else {
			valueField.SetBool(v)
		}
		break
	case reflect.String:
		valueField.SetString(strValue)
		break
	default:
		return fmt.Errorf("Unable to fill value of type [%s]", valueField.Type().Name())
	}
	return nil
}

func SetDisplayErrors(v bool) {
	displayErrors = v
}

func AnswerError(err error, code int, responseWriter http.ResponseWriter) {

	var msg string

	responseWriter.WriteHeader(code)

	if displayErrors {

		if err != nil {
			msg = err.Error()
		} else {
			msg = "An error occured"
		}

		responseWriter.Header().Add("Content-Type", TextPlainUTF8)

		_, err = responseWriter.Write([]byte(msg))
		if err != nil {
			appLog.DebugError(err, "Unable to write answer")
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func DisableClientCache(responseWriter http.ResponseWriter) {
	responseWriter.Header().Add("Cache-Control", "no-store")
}
