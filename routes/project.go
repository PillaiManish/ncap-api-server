package routes

import "net/http"

func (s *Server) initProjectRoutes() {
	s.router.HandleFunc("/create-new-project", s.apiServer.CreateNewProject).Methods(http.MethodPost)
	s.router.HandleFunc("/retrieve-project-list", s.apiServer.RetrieveProjectList).Methods(http.MethodGet)
	s.router.HandleFunc("/edit-project", s.apiServer.EditProject).Methods(http.MethodPatch)
	s.router.HandleFunc("/check-project-permission", s.apiServer.CheckProjectPermission).Methods(http.MethodPost)
	s.router.HandleFunc("/project-access-request", s.apiServer.ProjectAccessRequest).Methods(http.MethodPost)

}
