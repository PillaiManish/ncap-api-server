package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
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
	routes.HandleFunc("/drop-table-for-project", apiServer.CreateNewTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/insert-into-table-for-project", apiServer.InsertIntoTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/delete-table-for-project", apiServer.DropTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-table-list-for-project", apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
	//logger.Printf("Server started at port: %v", 8090)

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:4200"})

	http.ListenAndServe(":8090", handlers.CORS(headers, methods, origins)(routes))
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
	//w.Header().Set("Access-Control-Allow-Credentials", "true")
	//
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	//w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	//w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//
	//w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.CreateTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	ctx := context.Background()
	err = apiServer.database.CreateTableForProject(ctx, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.InsertIntoMetaData(ctx, tableModel)
	if err != nil {

		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) DropTableForProject(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.DropTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.DropTableForProject(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	retrieveTableListModel := &model.RetrieveTableListModel{}
	err = json.Unmarshal(body, retrieveTableListModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableList, err := apiServer.database.SelectTableListFromMetaData(context.Background(), retrieveTableListModel.ProjectId)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	resultStatus.Data = tableList
	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
	return
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

	ctx := context.Background()
	err = apiServer.database.InsertIntoTableForProject(ctx, tableModel)
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

func (apiServer *APIServerModel) RetrieveDataTableForProject(w http.ResponseWriter, req *http.Request) {
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
