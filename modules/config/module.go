package config

import (
	"fmt"
	"io/ioutil"

	appLog "github.com/dademo/rssreader/modules/log"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Feed struct {
	Name                 string `yaml:"name"`
	Url                  string `yaml:"url"`
	FetchIntervalMinutes uint   `yaml:"fetchIntervalMinutes"`
}

type DatabaseConfig struct {
	Driver             string `yaml:"driver"`
	ConnStr            string `yaml:"connStr"`
	MaxOpenConnections uint   `yaml:"maxOpenConnections"`
	MaxIdleConnexions  uint   `yaml:"maxIdleConnexions"`
}

type LogConfig struct {
	Level           string                 `yaml:"level"`
	Output          []string               `yaml:"output"`
	ReportCaller    bool                   `yaml:"reportCaller"`
	DisableColors   bool                   `yaml:"disableColors"`
	FullTimestamp   bool                   `yaml:"fullTimestamp"`
	TimestampFormat string                 `yaml:"timestampFormat"`
	Backends        *LogBackendsDefinition `yaml:"logBackendsDefinition"`
}

type HttpConfig struct {
	ListenAddress  string `yaml:"listenAddress"`
	StaticFilesDir string `yaml:"staticFilesDir"`
	DisplayErrors  bool   `yaml:"retrunErrors"`
}

type Config struct {
	Feeds      []*Feed         `yaml:"feeds"`
	DbConfig   *DatabaseConfig `yaml:"database"`
	LogConfig  *LogConfig      `yaml:"log"`
	HttpConfig *HttpConfig     `yaml:"http"`
}

func ReadConfig(configFilePath string) (*Config, error) {

	log.Debug(fmt.Sprintf("Reading config [%s]", configFilePath))

	fileContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		appLog.DebugError(err, "Unable to read the configuration file")
		return &Config{}, err
	}

	parsed := defaultConfig()
	err = yaml.Unmarshal(fileContent, parsed)
	if err != nil {
		appLog.DebugError(err, "Unable to unmarshall configuration")
		return &Config{}, err
	}

	log.Debug("Configuration have been read")

	return parsed, nil
}

func SaveConfig(config Config, configFilePath string) error {

	log.Debug("Marshalling configuration")

	marshalled, err := yaml.Marshal(config)
	if err != nil {
		appLog.DebugError(err, "Unable to marshall the configuration")
		return err
	}

	err = ioutil.WriteFile(configFilePath, marshalled, 0644)
	if err != nil {
		appLog.DebugError(err, "Unable to write the configuration")
		return err
	}

	return nil
}

func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           "info",
		Output:          []string{"stderr"},
		ReportCaller:    false,
		DisableColors:   false,
		FullTimestamp:   false,
		TimestampFormat: "",
		Backends:        defaultLogBackendsConfig(),
	}
}

func defaultConfig() *Config {
	return &Config{
		Feeds:      []*Feed{},
		DbConfig:   defaultDatabaseConfig(),
		LogConfig:  DefaultLogConfig(),
		HttpConfig: defaultHttpĈonfig(),
	}
}

func defaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:             "sqlite",
		ConnStr:            ":memory:",
		MaxOpenConnections: 10,
		MaxIdleConnexions:  5,
	}
}

func defaultHttpĈonfig() *HttpConfig {
	return &HttpConfig{
		ListenAddress:  "0.0.0.0",
		StaticFilesDir: "static",
		DisplayErrors:  true,
	}
}
