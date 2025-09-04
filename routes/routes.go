package routes

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"ncap-api-server/common"
	"ncap-api-server/controller"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type routesApp struct {
	router    *mux.Router
	apiServer controller.APIServer
}

func NewApiServer(ctx context.Context, logger *logrus.Logger, config *model.Config) error {
	newDB, err := database.NewDatabase(config.Database)
	if err != nil {
		logger.Fatalf("could not connect to database due to err: %s", err.Error())
	}
	if newDB == nil {
		logger.Fatalf("database connection return empty database")
	}

	apiServer := controller.NewApiServerModel(ctx, logger, newDB, config)

	routes := &routesApp{
		router:    mux.NewRouter(),
		apiServer: apiServer,
	}

	go func() {
		select {
		case <-ctx.Done():
			apiServer.Stop()
			return

		}
	}()

	routes.initAuthRoutes()
routes.initProjectRoutes()
routes.initTableRoutes()
routes.initTableDataRoutes()

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	allowedOrigins := []string{config.AdminPanelUrl}

	if os.Getenv(common.ADMIN_PANEL_URL) != "" {
		allowedOrigins = append(allowedOrigins, os.Getenv(common.ADMIN_PANEL_URL))
	}

	origins := handlers.AllowedOrigins(allowedOrigins)
	logger.Infof("ncap-api-server starting at port: %v", 8080)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(routes.router))
		if err != nil {
			logger.Fatalf("ncap-api-server failed to start due to err: %s", err.Error())
		}
	}()

	exitChan := SystemSignalHandling
	code := <-exitChan()

	logger.Debugf("Received signal to stop: %v", code)

	return nil
}

func SystemSignalHandling() chan int {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	exitChan := make(chan int, 1)

	select {
	case s := <-signalChan:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitChan <- 0
		default:
			println("unknown signal")
			exitChan <- 1
		}
	}
	return exitChan
}
