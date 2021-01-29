package config

import (
	"fmt"
	"io/ioutil"

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
	Level  string   `yaml:"level"`
	Output []string `yaml:"output"`
}

type HttpConfig struct {
	ListenAddress  string `yaml:"listenAddress"`
	StaticFilesDir string `yaml:"staticFilesDir"`
	DisplayErrors  bool   `yaml:"retrunErrors"`
}

type Config struct {
	Feeds      []Feed         `yaml:"feeds"`
	DbConfig   DatabaseConfig `yaml:"database"`
	LogConfig  LogConfig      `yaml:"log"`
	HttpConfig HttpConfig     `yaml:"http"`
}

func ReadConfig(configFilePath string) (Config, error) {

	log.Debug(fmt.Sprintf("Reading config [%s]", configFilePath))

	fileContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Error("Unable to read the configuration file")
		return Config{}, err
	}

	parsed := Config{}
	err = yaml.Unmarshal(fileContent, &parsed)
	if err != nil {
		log.Error("Unable to unmarshall configuration")
		return Config{}, err
	}

	log.Debug("Configuration have been read")

	return parsed, nil
}

func SaveConfig(config Config, configFilePath string) error {

	log.Debug("Marshalling configuration")

	marshalled, err := yaml.Marshal(config)
	if err != nil {
		log.Error("Unable to marshall the configuration")
		return err
	}

	err = ioutil.WriteFile(configFilePath, marshalled, 0644)
	if err != nil {
		log.Error("Unable to write the configuration")
		return err
	}

	return nil
}
