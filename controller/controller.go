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
	"ncap-api-server/common"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
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
	routes.HandleFunc("/create-table", apiServer.CreateNewTable).Methods(http.MethodPost)
	routes.HandleFunc("/drop-table", apiServer.DropTable).Methods(http.MethodPost)
	routes.HandleFunc("/insert-into-table", apiServer.InsertIntoTable).Methods(http.MethodPost)
	//routes.HandleFunc("/delete-table", apiServer.DropTable).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-table-list", apiServer.RetrieveTableListFromMetaData).Methods(http.MethodPost)
	routes.HandleFunc("/retrieve-data-from-table", apiServer.RetrieveDataTable).Methods(http.MethodPost)
	routes.HandleFunc("/delete-from-table", apiServer.DeleteEntryFromTable).Methods(http.MethodPost)
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
	w.Header().Set("Content-Type", common.JSONApplicationType)
	tokenString := req.Header.Get("Authorization")
	resultStatus := &model.ResultStatus{}

	if tokenString == "" {
		resultStatus.Error = fmt.Errorf("token not provided").Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	tokenString = tokenString[len("Bearer "):]

	userId, err := apiServer.verifyToken(tokenString)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	project := &model.ProjectModel{}
	err = json.Unmarshal(body, project)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return

	}

	project.UserId = userId

	err = apiServer.database.CreateNewProject(context.Background(), project, apiServer.config.ProjectConfig.Limit)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) RegisterUser(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newUser := &model.UserModel{}
	err = json.Unmarshal(body, newUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		resultStatus.Error = fmt.Errorf("name/password/email cannot be empty").Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	newUser.Password = string(encryptPasswd)

	err = apiServer.database.RegisterUser(context.Background(), newUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) LoginUser(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkUser := &model.UserModel{}
	err = json.Unmarshal(body, checkUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)

		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		resultStatus.Error = fmt.Errorf("email/password cannot be empty").Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	userId, err := apiServer.database.LoginUser(context.Background(), checkUser)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)

		return
	}

	token, err := apiServer.createToken(checkUser.Email, userId)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resultStatus)
	}
	resultStatus.Data = token

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)

}

func (apiServer *APIServerModel) CreateNewTable(w http.ResponseWriter, req *http.Request) {
	//w.Header().Set("Access-Control-Allow-Credentials", "true")
	//
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	//w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	//w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//
	//w.Header().Set("Content-Type", common.JSONApplicationType)
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.CreateTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	ctx := context.Background()
	err = apiServer.database.CreateTable(ctx, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.InsertIntoMetaData(ctx, tableModel)
	if err != nil {

		resultStatus.Error = err.Error()
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

func (apiServer *APIServerModel) DropTable(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableModel := &model.DropTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	err = apiServer.database.DropTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
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
	body, err := io.ReadAll(req.Body)
	fmt.Printf("Received %s", string(body))
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	retrieveTableListModel := &model.RetrieveTableListModel{}
	err = json.Unmarshal(body, retrieveTableListModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}

	tableList, err := apiServer.database.SelectTableListFromMetaData(context.Background(), retrieveTableListModel.ProjectId)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(resultStatus)
		return
	}
	resultStatus.Data = tableList
	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (apiServer *APIServerModel) InsertIntoTable(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.InsertIntoTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	ctx := context.Background()
	err = apiServer.database.InsertIntoTable(ctx, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resultStatus.Status = common.SUCCESS
	w.Header().Set("Content-Type", common.JSONApplicationType)
	json.NewEncoder(w).Encode(resultStatus)
}

func (apiServer *APIServerModel) RetrieveDataTable(w http.ResponseWriter, req *http.Request) {
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.RetrieveDataFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	columns, data, err := apiServer.database.RetrieveDataFromTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
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
	//apiServer.logger.Printf("Received request: %s", req.Body)
	body, err := io.ReadAll(req.Body)
	resultStatus := &model.ResultStatus{}

	if err != nil {
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	tableModel := &model.DeleteEntryFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = apiServer.database.DeleteEntryFromTable(context.Background(), tableModel)
	if err != nil {
		resultStatus.Error = err.Error()
		resultStatus.Status = common.FAILED
		//apiServer.logger.Printf("common.FAILED to read body due to err: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	resultStatus.Status = common.SUCCESS
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
