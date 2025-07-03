package routes

import "net/http"

func (routesApp *routesApp) initAuthRoutes() {
	routesApp.router.HandleFunc("/login", routesApp.apiServer.Login).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/register", routesApp.apiServer.Register).Methods(http.MethodPost)
}
