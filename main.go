package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"ncap-api-server/model"
	"ncap-api-server/routes"
	"os"
)

func main() {
	var err error
	logger := logrus.New()
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

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

	err = routes.NewApiServer(ctx, logger, config)
	if err != nil {
		logger.Fatal(err)
	}

}
