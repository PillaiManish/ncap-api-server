package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"ncap-api-server/common"
	"ncap-api-server/controller"
	"ncap-api-server/database"
	"ncap-api-server/model"
)

const (
	serverPort      = ":8080"
	shutdownTimeout = 30 * time.Second
	writeTimeout    = 15 * time.Second
	readTimeout     = 15 * time.Second
	idleTimeout     = 60 * time.Second
)

// Server represents the HTTP server and its dependencies
type Server struct {
	router    *mux.Router
	apiServer controller.Controller
	logger    *logrus.Logger
	config    *model.Config
	server    *http.Server
}

// NewApiServer creates and starts a new API server
func NewApiServer(ctx context.Context, logger *logrus.Logger, config *model.Config) error {
	// Initialize database
	db, err := database.NewDatabase(config.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	if db == nil {
		return fmt.Errorf("database connection returned nil")
	}

	// Create a server instance
	server := &Server{
		router:    mux.NewRouter(),
		apiServer: controller.NewApiServerModel(ctx, logger, db, config),
		logger:    logger,
		config:    config,
	}

	// Setup server
	server.setupRoutes()
	server.setupCORS()
	server.createHTTPServer()

	// Start server with graceful shutdown
	return server.startWithGracefulShutdown(ctx)
}

// setupRoutes initializes all API routes
func (s *Server) setupRoutes() {
	s.logger.Debug("Setting up API routes")

	s.initAuthRoutes()
	s.initProjectRoutes()
	s.initTableRoutes()
	s.initTableDataRoutes()

	s.logger.Debug("API routes setup completed")
}

// setupCORS configures CORS middleware
func (s *Server) setupCORS() {
	// CORS is now applied at the server level in createHTTPServer()
	// This function is kept for future CORS-related setup if needed
	s.logger.Debug("CORS configuration will be applied at server level")
}

// getCORSHeaders returns the allowed CORS headers
func (s *Server) getCORSHeaders() handlers.CORSOption {
	return handlers.AllowedHeaders([]string{
		"Accept",
		"Accept-Language",
		"Content-Language",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Origin",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
	})
}

// getCORSMethods returns the allowed CORS methods
func (s *Server) getCORSMethods() handlers.CORSOption {
	return handlers.AllowedMethods([]string{
		"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
	})
}

// getCORSOrigins returns the allowed CORS origins
func (s *Server) getCORSOrigins() handlers.CORSOption {
	origins := []string{s.config.AdminPanelUrl}

	if envOrigin := os.Getenv(common.ADMIN_PANEL_URL); envOrigin != "" {
		origins = append(origins, envOrigin)
	}

	return handlers.AllowedOrigins(origins)
}

// createHTTPServer creates the HTTP server instance
func (s *Server) createHTTPServer() {
	// Apply CORS at the server level
	corsHandler := handlers.CORS(
		s.getCORSHeaders(),
		s.getCORSMethods(),
		s.getCORSOrigins(),
	)(s.router)

	s.server = &http.Server{
		Addr:         serverPort,
		Handler:      corsHandler,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}
}

// startWithGracefulShutdown starts the server with graceful shutdown handling
func (s *Server) startWithGracefulShutdown(ctx context.Context) error {
	// Create a shutdown signal channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go s.startServer(serverErr)

	// Wait for a shutdown signal or server error
	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdownChan:
		s.logger.Infof("Received shutdown signal: %v", sig)
		return s.gracefulShutdown()
	case <-ctx.Done():
		s.logger.Info("Context cancelled, shutting down server")
		return s.gracefulShutdown()
	}
}

// startServer starts the HTTP server
func (s *Server) startServer(serverErr chan<- error) {
	s.logger.Infof("Starting ncap-api-server on port %s", serverPort)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		serverErr <- fmt.Errorf("server failed to start: %w", err)
	}
}

// gracefulShutdown performs a graceful shutdown of the server
func (s *Server) gracefulShutdown() error {
	s.logger.Info("Starting graceful shutdown...")

	// Stop the API server first
	s.apiServer.Stop()

	// Create a shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Errorf("Server forced to shutdown: %v", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("Server shutdown completed successfully")
	return nil
}
