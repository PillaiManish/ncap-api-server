package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
	"time"
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
		config   *model.Config
	}
)

//var secretKey = []byte("secret-key")

func NewApiServer(logger *log.Logger, config *model.Config) error {
	apiServer := &APIServerModel{logger: logger, config: config}

	routes := mux.NewRouter()

	newDB, err := database.NewDatabase(config.Database)
	if err != nil {
		return err
	}
	apiServer.database = newDB

	// Authentication
	routes.HandleFunc("/login", apiServer.LoginUser).Methods(http.MethodPost)
	routes.HandleFunc("/register", apiServer.RegisterUser).Methods(http.MethodPost)

	routes.HandleFunc("/create-new-project", apiServer.CreateNewProject).Methods(http.MethodPost)
	routes.HandleFunc("/create-table-for-project", apiServer.CreateNewTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/drop-table-for-project", apiServer.CreateNewTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/insert-into-table-for-project", apiServer.InsertIntoTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/delete-table-for-project", apiServer.DropTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-table-list-for-project", apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-data-from-table-for-project", apiServer.RetrieveDataTableForProject).Methods(http.MethodPost)
	routes.HandleFunc("/delete-from-table-for-project", apiServer.DeleteEntryFromTable).Methods(http.MethodPost)
	//logger.Printf("Server started at port: %v", 8090)

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:4200"})

	err = http.ListenAndServe(":8090", handlers.CORS(headers, methods, origins)(routes))
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

func (apiServer *APIServerModel) RegisterUser(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newUser := &model.UserModel{}
	err = json.Unmarshal(body, newUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		resultStatus.Error = fmt.Errorf("name/password/email cannot be empty").Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	newUser.Password = string(encryptPasswd)

	err = apiServer.database.RegisterUser(context.Background(), newUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) LoginUser(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkUser := &model.UserModel{}
	err = json.Unmarshal(body, checkUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		resultStatus.Error = fmt.Errorf("email/password cannot be empty").Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.LoginUser(context.Background(), checkUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	token, err := apiServer.createToken(checkUser.Email)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)
	}
	resultStatus.Data = token

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

	tableModel := &model.RetrieveDataFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	columns, data, err := apiServer.database.RetrieveDataFromTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	resultStatus.Data = map[string]interface{}{
		"columns": columns,
		"rows":    data,
	}

	resultStatus.Status = SUCCESS
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) DeleteEntryFromTable(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.DeleteEntryFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = FAILED
		//apiServer.logger.Printf("failed to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.DeleteEntryFromTable(context.Background(), tableModel)
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

func (apiServer *APIServerModel) createToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(apiServer.config.JWT.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (apiServer *APIServerModel) verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return apiServer.config.JWT.Secret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
