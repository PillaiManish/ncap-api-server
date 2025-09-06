package routes

import "net/http"

func (routesApp *routesApp) initProjectRoutes() {
	routesApp.router.HandleFunc("/create-new-project", routesApp.apiServer.CreateNewProject).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/retrieve-project-list", routesApp.apiServer.RetrieveProjectList).Methods(http.MethodGet)
	routesApp.router.HandleFunc("/edit-project", routesApp.apiServer.EditProject).Methods(http.MethodPatch)
	routesApp.router.HandleFunc("/check-project-permission", routesApp.apiServer.CheckProjectPermission).Methods(http.MethodPost)
	routesApp.router.HandleFunc("/project-access-request", routesApp.apiServer.ProjectAccessRequest).Methods(http.MethodPost)

}
