package routes

import "net/http"

func (s *Server) initTableRoutes() {
	s.router.HandleFunc("/create-table", s.apiServer.CreateNewTable).Methods(http.MethodPost)
	s.router.HandleFunc("/drop-table", s.apiServer.DropTable).Methods(http.MethodPost)
	s.router.HandleFunc("/retrieve-data-from-table", s.apiServer.RetrieveDataTable).Methods(http.MethodPost)

}
