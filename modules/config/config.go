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
	driver             string
	connStr            string
	maxOpenConnections uint
	maxIdleConnexions  uint
}

type LogConfig struct {
	level  string
	output []string
}

type Config struct {
	feeds []Feed
	db    DatabaseConfig
	log   LogConfig
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
		log.Fatal("Unable to marshall the configuration")
		return err
	}

	err = ioutil.WriteFile(configFilePath, marshalled, 0644)
	if err != nil {
		log.Fatal("Unable to write the configuration")
		return err
	}

	return nil
}
