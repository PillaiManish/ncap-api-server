package main

import (
	"log"
	"ncap-api-server/controller"
)

func main() {

	logger := &log.Logger{}

	err := controller.NewApiServer(logger)
	if err != nil {
		logger.Fatal(err)
	}

}
