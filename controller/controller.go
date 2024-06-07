package controller

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
)

const (
	FAILED  = "FAILED"
	SUCCESS = "SUCCESS"
)

type (
	APIServer interface {
	}

	APIServerModel struct {
		logger   *log.Logger
		database database.Database
	}
)

func NewApiServer(logger *log.Logger) error {
	apiServer := &APIServerModel{logger: logger}

	routes := mux.NewRouter()

	newDB, err := database.NewDatabase("MySQL")
	if err != nil {
		return err
	}
	apiServer.database = newDB

	routes.HandleFunc("/create-new-project", apiServer.CreateNewProject).Methods(http.MethodPost)
	routes.HandleFunc("/create-table-for-project", apiServer.CreateNewTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/insert-into-table-for-project", apiServer.InsertIntoTableForProject).Methods(http.MethodPost)

	//logger.Printf("Server started at port: %v", 8090)

	err = http.ListenAndServe(":8090", routes)
	if err != nil {
		return err
	}

	return nil
}

func (apiServer *APIServerModel) CreateNewProject(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	project := &model.ProjectModel{}
	err = json.Unmarshal(body, project)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.CreateNewProject(context.Background(), project.Id, project.Name, project.Desc)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) CreateNewTableForProject(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.CreateTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.CreateTableForProject(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) InsertIntoTableForProject(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.InsertIntoTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.InsertIntoTableForProject(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
}
