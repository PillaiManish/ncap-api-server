package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"ncap-api-server/controller"
	"ncap-api-server/model"
	"os"
)

func main() {
	var err error
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		panic(err)
	}

	config := &model.Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		panic(err)
	}

	err = controller.NewApiServer(logger, config)
	if err != nil {
		logger.Fatal(err)
	}

}
