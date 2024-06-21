package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"ncap-api-server/controller"
	"ncap-api-server/model"
	"os"
)

func main() {
	var err error
	logger := logrus.New()
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		panic(err)
	}

	config := &model.Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}

	logLevel := logrus.InfoLevel

	if level, err := logrus.ParseLevel(config.LogLevel); err == nil {
		logLevel = level
	}
	logger.SetLevel(logLevel)

	err = controller.NewApiServer(logger, config)
	if err != nil {
		logger.Fatal(err)
	}

}
