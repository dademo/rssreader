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

	"github.com/dademo/rssreader/modules/config"
	appLog "github.com/dademo/rssreader/modules/log"
)

type RegisteredRoute struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

var (
	registeredRoutes []RegisteredRoute
)

const (
	HTTPParameterTagName         = "httpParameter"
	HTTPOptionalParameterTagName = "httpOptionalParameter"
	JSONContentTypeUtf8          = "application/json; charset=utf-8"
	headerContentType            = "Content-Type"
	contentTypeMultipartFormData = "multipart/form-data"
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

	responseWriter.Header().Add("Content-Type", JSONContentTypeUtf8)

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

		strURLValue := request.Form.Get(urlParameter)

		if strURLValue != "" {
			switch valueField.Type().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if v, err := strconv.ParseInt(strURLValue, 10, 0); err != nil {
					return err
				} else {
					valueField.SetInt(v)
				}
				break
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if v, err := strconv.ParseUint(strURLValue, 10, 0); err != nil {
					return err
				} else {
					valueField.SetUint(v)
				}
				break
			case reflect.Float32, reflect.Float64:
				if v, err := strconv.ParseFloat(strURLValue, 0); err != nil {
					return err
				} else {
					valueField.SetFloat(v)
				}
				break
			case reflect.Bool:
				if v, err := strconv.ParseBool(strURLValue); err != nil {
					return err
				} else {
					valueField.SetBool(v)
				}
				break
			case reflect.String:
				valueField.SetString(strURLValue)
				break
			default:
				return fmt.Errorf("Unable to fill value of type [%s]", valueField.Type().Name())
			}
		} else if _, ok := tag.Lookup(HTTPOptionalParameterTagName); !ok {
			return fmt.Errorf("Non-optional field [%s] (http parameter '%s') is missing", typeField.Name, urlParameter)
		}
	}

	return nil
}
