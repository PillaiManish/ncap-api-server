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

	ResultStatus struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
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

	routes.HandleFunc("/create-new-project/", apiServer.CreateNewProject).Methods(http.MethodPost)

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
	resultStatus := &ResultStatus{}

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
		return
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)

}
