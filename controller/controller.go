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
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
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
		ctx      context.Context
	}
)

//var secretKey = []byte("secret-key")

func NewApiServer(ctx context.Context, logger *logrus.Logger, config *model.Config) error {
	apiServer := &APIServerModel{logger: logger, config: config, ctx: ctx}

	routes := mux.NewRouter()

	newDB, err := database.NewDatabase(config.Database)
	if err != nil {
		logger.Fatalf("could not connect to database due to err: %s", err.Error())
	}

	if newDB == nil {
		logger.Fatalf("database connection return empty database")
	}
	apiServer.database = newDB

	go func() {
		select {
		case <-ctx.Done():
			logger.Debugf("closing database connection")
			apiServer.database.Close()
			return

		}
	}()

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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(routes))
		//err = http.ListenAndServe(":8080", routes)
		if err != nil {
			logger.Fatalf("ncap-api-server failed to start due to err: %s", err.Error())
		}
	}()

	exitChan := SystemSignalHandling
	code := <-exitChan()

	logger.Debugf("Received signal to stop: %v", code)
	apiServer.Stop()

	wg.Wait()

	return nil
}

func SystemSignalHandling() chan int {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	exitChan := make(chan int, 1)

	select {
	case s := <-signalChan:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitChan <- 0
		default:
			println("unknown signal")
			exitChan <- 1
		}
	}
	return exitChan
}

func (apiServer *APIServerModel) Stop() {
	apiServer.database.Close()
	apiServer.logger.Debugf("db closed and closed api-server")
	apiServer.logger.Exit(1)
}

func (apiServer *APIServerModel) Response(w http.ResponseWriter, code int, payload interface{}) {
	resultStatus := &model.ResultStatus{}

	switch code {
	case http.StatusOK:
		resultStatus.Status = common.SUCCESS
		resultStatus.StatusCode = http.StatusOK
		resultStatus.Message = "Processed successfully"
		resultStatus.Data = payload
		w.WriteHeader(http.StatusOK)
	case http.StatusBadRequest:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusBadRequest
		resultStatus.Message = "Invalid request"

		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusBadRequest)
	case http.StatusUnauthorized:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		resultStatus.Message = "Unauthorized"
		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusUnauthorized)
	case http.StatusNotFound:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusNotFound
		resultStatus.Message = "Not found"
		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusNotFound)
	case http.StatusInternalServerError:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		resultStatus.Message = "Internal server error"
		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusInternalServerError)
	case http.StatusConflict:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusConflict
		resultStatus.Message = "Conflict"
		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusConflict)
	case http.StatusCreated:
		resultStatus.Status = common.SUCCESS
		resultStatus.StatusCode = http.StatusCreated
		resultStatus.Message = "Processed successfully"
		resultStatus.Data = payload
		w.WriteHeader(http.StatusCreated)
	default:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		resultStatus.Message = "Internal server error"
		if apiServer.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(resultStatus)
	return
}
func (apiServer *APIServerModel) CreateNewProject(w http.ResponseWriter, req *http.Request) {
	endpointName := "Create New Project"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	project := &model.ProjectModel{}
	err = json.Unmarshal(body, project)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	project.UserId = userId

	statusCode, err := apiServer.database.CreateNewProject(apiServer.ctx, project, apiServer.config.ProjectConfig.Limit)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, statusCode, nil)
	return

}

func (apiServer *APIServerModel) RegisterUser(w http.ResponseWriter, req *http.Request) {
	endpointName := "Register"
	apiServer.logger.Debugf("Received %s request", endpointName)
	body, err := io.ReadAll(req.Body)

	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	newUser := &model.UserModel{}
	err = json.Unmarshal(body, newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		err := fmt.Errorf("name/password/email cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 1)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser.Password = string(encryptPasswd)

	userId, statusCode, err := apiServer.database.RegisterUser(apiServer.ctx, newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	newUser.Id = int(userId)

	token, err := apiServer.createToken(newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusCreated, token)
	return
}

func (apiServer *APIServerModel) LoginUser(w http.ResponseWriter, req *http.Request) {
	endpointName := "Login"
	apiServer.logger.Debugf("Received %s request", endpointName)
	body, err := io.ReadAll(req.Body)

	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	checkUser := &model.UserModel{}
	err = json.Unmarshal(body, checkUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		err := fmt.Errorf("email/password cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	userModel, statusCode, err := apiServer.database.LoginUser(apiServer.ctx, checkUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	token, err := apiServer.createToken(userModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, token)
	return
}

func (apiServer *APIServerModel) CreateNewTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Create-New-Table"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
	}

	params := req.URL.Query()
	paramsProjectId := params.Get("project_id")

	projectId, err := strconv.Atoi(paramsProjectId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	tableModel := &model.CreateTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	tableModel.TableName = CreateTableNameFromName(tableModel.DisplayName, projectId)

	statusCode, err := apiServer.database.CreateTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	statusCode, err = apiServer.database.InsertIntoMetaData(apiServer.ctx, tableModel, projectId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusCreated, nil)

	return
}

func (apiServer *APIServerModel) DropTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Drop Table"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	tableModel := &model.DropTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = apiServer.database.DropTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
	return
}

func (apiServer *APIServerModel) RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Table List"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	retrieveTableListModel := &model.RetrieveTableListModel{}
	err = json.Unmarshal(body, retrieveTableListModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if retrieveTableListModel.ProjectId == "" {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return

	}

	tableList, statusCode, err := apiServer.database.RetrieveTableListFromMetaData(apiServer.ctx, retrieveTableListModel.ProjectId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, tableList)
	return
}

func (apiServer *APIServerModel) InsertIntoTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Insert Into Table"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	tableModel := &model.InsertIntoTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	err = apiServer.database.InsertIntoTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
	return
}

func (apiServer *APIServerModel) RetrieveDataTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Data Table List"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	tableModel := &model.RetrieveDataFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	columns, data, err := apiServer.database.RetrieveDataFromTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
	}

	resultData := map[string]interface{}{
		"columns": columns,
		"rows":    data,
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, resultData)
	return
}

func (apiServer *APIServerModel) DeleteEntryFromTable(w http.ResponseWriter, req *http.Request) {
	endpointName := " Delete Entry"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	_, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	tableModel := &model.DeleteEntryFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
	}

	err = apiServer.database.DeleteEntryFromTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
}

func (apiServer *APIServerModel) Authorization(w http.ResponseWriter, authorizationHeader, endpointName string) (int, error) {

	if authorizationHeader == "" {
		err := fmt.Errorf("missing Authorization header")
		return -1, err
	}

	tokenString := authorizationHeader[len("Bearer "):]

	userId, err := apiServer.verifyToken(tokenString)
	if err != nil {
		return -1, err
	}

	return userId, nil
}
func (apiServer *APIServerModel) RetrieveProjectList(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Project List"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := apiServer.Authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	projects, statusCode, err := apiServer.database.RetrieveProjectList(apiServer.ctx, userId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return

	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, projects)
	return
}

func (apiServer *APIServerModel) createToken(userModel *model.UserModel) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":     userModel.Id,
			"name":   userModel.Name,
			"email":  userModel.Email,
			"userId": userModel.Id,
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
