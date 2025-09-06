package routes

import "net/http"

func (routesApp *routesApp) initTableRoutes() {
	routesApp.router.HandleFunc("/create-table", routesApp.apiServer.CreateNewTable).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/drop-table", routesApp.apiServer.DropTable).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/retrieve-data-from-table", routesApp.apiServer.RetrieveDataTable).Methods(http.MethodPost)

}
