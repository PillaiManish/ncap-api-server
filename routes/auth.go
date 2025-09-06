package routes

import "net/http"

func (s *Server) initAuthRoutes() {
	s.router.HandleFunc("/login", s.apiServer.Login).Methods(http.MethodPost)
	s.router.HandleFunc("/register", s.apiServer.Register).Methods(http.MethodPost)
}
