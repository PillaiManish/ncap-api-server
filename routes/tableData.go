package routes

import "net/http"

func (s *Server) initTableDataRoutes() {
	s.router.HandleFunc("/insert-into-table", s.apiServer.InsertIntoTable).Methods(http.MethodPost)
	s.router.HandleFunc("/delete-from-table", s.apiServer.DeleteEntryFromTable).Methods(http.MethodPost)
	s.router.HandleFunc("/retrieve-table-list", s.apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
	s.router.HandleFunc("/update-table-data", s.apiServer.UpdateTableData).Methods(http.MethodPatch)
}
