package config

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Feed struct {
	name                 string
	url                  string
	fetchIntervalMinutes uint
}

type DatabaseConfig struct {
	driver  string
	connStr string
}

type Config struct {
	feeds []Feed
	db    DatabaseConfig
}

func ReadConfig(configFilePath string) Config {

	log.Debug(fmt.Sprintf("Reading config [%s]", configFilePath))

	fileContent, err := ioutil.ReadFile(configFilePath)
	check(err, "Unable to read the configuration file")

	parsed := Config{}
	err = yaml.Unmarshal(fileContent, &parsed)
	check(err, "Unable to unmarshall configuration")

	log.Debug("Configuration have been read")

	return parsed
}

func SaveConfig(config Config, configFilePath string) {

	log.Debug("Marshalling configuration")

	marshalled, err := yaml.Marshal(config)

	check(err, "Unable to marshall the configuration")

	err = ioutil.WriteFile(configFilePath, marshalled, 0644)

	check(err, "Unable to write the configuration")
}

func check(err error, msg string) {
	if err != nil {
		log.Fatal(msg, ", ", err)
	}
}
