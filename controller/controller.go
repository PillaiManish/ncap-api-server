package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"ncap-api-server/common"
	"ncap-api-server/database"
	"ncap-api-server/model"
	"net/http"
	"strings"
	"time"
)

type JwtClaims struct {
	UserID   int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`

	jwt.RegisteredClaims
	//jwt.StandardClaims
}

type (
	APIServer interface {
		Login(w http.ResponseWriter, r *http.Request)
		Register(w http.ResponseWriter, r *http.Request)

		CreateNewTable(w http.ResponseWriter, req *http.Request)
		DropTable(w http.ResponseWriter, req *http.Request)
		RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request)

		CreateNewProject(w http.ResponseWriter, req *http.Request)
		EditProject(w http.ResponseWriter, req *http.Request)
		CheckProjectPermission(w http.ResponseWriter, req *http.Request)
		ProjectAccessRequest(w http.ResponseWriter, req *http.Request)
		RetrieveProjectList(w http.ResponseWriter, req *http.Request)

		InsertIntoTable(w http.ResponseWriter, req *http.Request)
		RetrieveDataTable(w http.ResponseWriter, req *http.Request)
		DeleteEntryFromTable(w http.ResponseWriter, req *http.Request)

		Stop()
	}

	APIServerModel struct {
		logger   *logrus.Logger
		database database.Database
		config   *model.Config
		ctx      context.Context
	}
)

func NewApiServerModel(ctx context.Context, logger *logrus.Logger, database database.Database, config *model.Config) APIServer {
	return &APIServerModel{
		logger:   logger,
		database: database,
		config:   config,
		ctx:      ctx,
	}
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

func (apiServer *APIServerModel) authorization(w http.ResponseWriter, authorizationHeader, endpointName string) (int, error) {

	if authorizationHeader == "" {
		err := fmt.Errorf("missing authorization header")
		return -1, err
	}

	tokenString := authorizationHeader[len("Bearer "):]

	userId, err := apiServer.verifyToken(tokenString)
	if err != nil {
		return -1, err
	}

	return userId, nil
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

func (apiServer *APIServerModel) createToken(userModel *model.UserModel) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":    userModel.Id,
			"name":  userModel.Name,
			"email": userModel.Email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(apiServer.config.JWT.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (apiServer *APIServerModel) checkPermission(userId int, projectId string, permission common.PERMISSION) (bool, error) {
	apiServer.logger.Debugf("Checking if user %d has permission to project %s", userId, projectId)
	res, err := apiServer.database.CheckPermission(apiServer.ctx, userId, projectId, permission)
	if err != nil {
		return false, err
	}

	if res == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func (apiServer *APIServerModel) createTableNameFromName(displayName string, projectId int) string {
	tableName := strings.ReplaceAll(displayName, "-", "_")
	tableName = strings.ReplaceAll(tableName, " ", "_")

	return fmt.Sprintf("%v_%s", projectId, tableName)
}

func (apiServer *APIServerModel) Stop() {
	apiServer.database.Close()
	apiServer.logger.Debugf("db closed and closed api-server")
	apiServer.logger.Exit(1)
}
