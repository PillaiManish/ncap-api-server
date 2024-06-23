package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"io"
	"ncap-api-server/common"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type JwtClaims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
	//jwt.StandardClaims
}

type (
	APIServer interface {
	}

	APIServerModel struct {
		logger   *logrus.Logger
		database database.Database
		config   *model.Config
	}
)

//var secretKey = []byte("secret-key")

func NewApiServer(logger *logrus.Logger, config *model.Config) error {
	apiServer := &APIServerModel{logger: logger, config: config}

	routes := mux.NewRouter()

	newDB, err := database.NewDatabase(config.Database)
	if err != nil {
		logger.Fatalf("could not connect to database due to err: %s", err.Error())
	}

	if newDB == nil {
		logger.Fatalf("database connection return empty database")
	}
	apiServer.database = newDB

	// Authentication
	routes.HandleFunc("/login", apiServer.LoginUser).Methods(http.MethodPost)
	routes.HandleFunc("/register", apiServer.RegisterUser).Methods(http.MethodPost)

	routes.HandleFunc("/create-new-project", apiServer.CreateNewProject).Methods(http.MethodPost)
	routes.HandleFunc("/create-table", apiServer.CreateNewTable).Methods(http.MethodPost)
	routes.HandleFunc("/drop-table", apiServer.DropTable).Methods(http.MethodPost)
	routes.HandleFunc("/insert-into-table", apiServer.InsertIntoTable).Methods(http.MethodPost)
	//routes.HandleFunc("/delete-table", apiServer.DropTable).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-table-list", apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-data-from-table", apiServer.RetrieveDataTable).Methods(http.MethodPost)
	routes.HandleFunc("/delete-from-table", apiServer.DeleteEntryFromTable).Methods(http.MethodPost)

	// TODO: list of projects
	routes.HandleFunc("/retrieve-project-list", apiServer.RetrieveProjectList).Methods(http.MethodGet)

	//logger.Printf("Server started at port: %v", 8090)

	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	allowedOrigins := []string{config.AdminPanelUrl}

	if os.Getenv(common.ADMIN_PANEL_URL) != "" {
		allowedOrigins = append(allowedOrigins, os.Getenv(common.ADMIN_PANEL_URL))
	}

	origins := handlers.AllowedOrigins(allowedOrigins)
	logger.Infof("ncap-api-server starting at port: %v", 8080)
	err = http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(routes))
	//err = http.ListenAndServe(":8080", routes)
	if err != nil {
		logger.Fatalf("ncap-api-server failed to start due to err: %s", err.Error())
	}

	return nil
}

func (apiServer *APIServerModel) CreateNewProject(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Create-New-Project request")
	w.Header().Set("Content-Type", common.JSONApplicationType)
	resultStatus := &model.ResultStatus{}

	w.WriteHeader(http.StatusOK)

	tokenString := req.Header.Get("Authorization")
	if tokenString == "" {
		resultStatus.Message = fmt.Errorf("token not provided").Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	tokenString = tokenString[len("Bearer "):]

	userId, err := apiServer.verifyToken(tokenString)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	project := &model.ProjectModel{}
	err = json.Unmarshal(body, project)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return

	}

	project.UserId = userId

	statusCode, err := apiServer.database.CreateNewProject(context.Background(), project, apiServer.config.ProjectConfig.Limit)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = common.SUCCESS
	resultStatus.StatusCode = http.StatusOK
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) RegisterUser(w http.ResponseWriter, req *http.Request) {
	endpointName := "Login"
	apiServer.logger.Debugf("Received %s request", endpointName)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}
	w.WriteHeader(http.StatusOK)

	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newUser := &model.UserModel{}
	err = json.Unmarshal(body, newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		err := fmt.Errorf("name/password/email cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	newUser.Password = string(encryptPasswd)

	userId, statusCode, err := apiServer.database.RegisterUser(context.Background(), newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	token, err := apiServer.createToken(newUser.Email, int(userId))
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		json.NewEncoder(w).Encode(resultStatus)
	}
	resultStatus.Data = token
	resultStatus.Status = common.SUCCESS
	resultStatus.StatusCode = http.StatusCreated
	apiServer.logger.Debugf("%s: successfully processed", endpointName)

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) LoginUser(w http.ResponseWriter, req *http.Request) {
	endpointName := "Login"
	apiServer.logger.Debugf("Received %s request", endpointName)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	w.WriteHeader(http.StatusOK)

	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)

		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkUser := &model.UserModel{}
	err = json.Unmarshal(body, checkUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		err := fmt.Errorf("email/password cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	userId, statusCode, err := apiServer.database.LoginUser(context.Background(), checkUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	token, err := apiServer.createToken(checkUser.Email, userId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		json.NewEncoder(w).Encode(resultStatus)
	}
	resultStatus.Data = token
	resultStatus.Status = common.SUCCESS
	resultStatus.StatusCode = http.StatusOK

	apiServer.logger.Debugf("%s: successfully processed", endpointName)

	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) CreateNewTable(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Create-New-Table request")
	w.WriteHeader(http.StatusOK)

	body, err := io.ReadAll(req.Body)
	//fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	params := req.URL.Query()
	project_id := params.Get("project_id")

	projectId, err := strconv.Atoi(project_id)
	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.CreateTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel.TableName = CreateTableNameFromName(tableModel.DisplayName, projectId)

	ctx := context.Background()
	statusCode, err := apiServer.database.CreateTable(ctx, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	statusCode, err = apiServer.database.InsertIntoMetaData(ctx, tableModel, projectId)
	if err != nil {

		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = common.SUCCESS
	resultStatus.StatusCode = http.StatusCreated
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) DropTable(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Drop-Table request")
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.DropTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.DropTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request) {
	endpoint := "Retrieve - Table - List"
	apiServer.logger.Debugf("Received %s request", endpoint)

	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	retrieveTableListModel := &model.RetrieveTableListModel{}
	err = json.Unmarshal(body, retrieveTableListModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	if retrieveTableListModel.ProjectId == "" {
		resultStatus.Message = fmt.Errorf("project id cannot be null").Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		json.NewEncoder(w).Encode(resultStatus)
		return

	}

	tableList, statusCode, err := apiServer.database.SelectTableListFromMetaData(context.Background(), retrieveTableListModel.ProjectId)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	resultStatus.Data = tableList
	resultStatus.Status = common.SUCCESS
	resultStatus.StatusCode = http.StatusOK
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) InsertIntoTable(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Insert-Into-Table request")
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.InsertIntoTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	ctx := context.Background()
	err = apiServer.database.InsertIntoTable(ctx, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) RetrieveDataTable(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Data Table request")
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.RetrieveDataFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	columns, data, err := apiServer.database.RetrieveDataFromTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	resultStatus.Data = map[string]interface{}{
		"columns": columns,
		"rows":    data,
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) DeleteEntryFromTable(w http.ResponseWriter, req *http.Request) {
	apiServer.logger.Debugf("Received Delete-Entry request")
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.DeleteEntryFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.DeleteEntryFromTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) RetrieveProjectList(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Project List"
	apiServer.logger.Debugf("Received %s request", endpointName)

	resultStatus := &model.ResultStatus{}
	w.WriteHeader(http.StatusOK)

	tokenString := req.Header.Get("Authorization")
	if tokenString == "" {
		resultStatus.Message = fmt.Errorf("token not provided").Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	tokenString = tokenString[len("Bearer "):]

	userId, err := apiServer.verifyToken(tokenString)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	projects, statusCode, err := apiServer.database.RetrieveProjectList(context.Background(), userId)
	if err != nil {
		resultStatus.Message = err.Error()
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = statusCode
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return

	}
	resultStatus.Status = common.SUCCESS
	resultStatus.Data = projects
	resultStatus.StatusCode = http.StatusOK

	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) createToken(email string, userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email":  email,
			"userId": userId,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(apiServer.config.JWT.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (apiServer *APIServerModel) verifyToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return apiServer.config.JWT.Secret, nil
	})

	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, fmt.Errorf("invalid token")
	}
	userId := -1
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		userId = claims.UserID
		return userId, nil

	} else {
		return -1, fmt.Errorf("invalid token")
	}
}

func CreateTableNameFromName(displayName string, projectId int) string {
	tableName := strings.ReplaceAll(displayName, "-", "_")
	tableName = strings.ReplaceAll(tableName, " ", "_")

	return fmt.Sprintf("%v_%s", projectId, tableName)
}
