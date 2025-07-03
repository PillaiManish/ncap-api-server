package routes

import "net/http"

func (routesApp *routesApp) initTableDataRoutes() {
	routesApp.router.HandleFunc("/insert-into-table", routesApp.apiServer.InsertIntoTable).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/delete-from-table", routesApp.apiServer.DeleteEntryFromTable).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/retrieve-table-list", routesApp.apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
}
