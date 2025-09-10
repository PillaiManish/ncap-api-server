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
	Controller interface {
		Login(w http.ResponseWriter, r *http.Request)
		Register(w http.ResponseWriter, r *http.Request)

		CreateNewTable(w http.ResponseWriter, req *http.Request)
		DropTable(w http.ResponseWriter, req *http.Request)
		RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request)
		UpdateTableData(w http.ResponseWriter, req *http.Request)

		CreateNewProject(w http.ResponseWriter, req *http.Request)
		EditProject(w http.ResponseWriter, req *http.Request)
		CheckProjectPermission(w http.ResponseWriter, req *http.Request)
		ProjectAccessRequest(w http.ResponseWriter, req *http.Request)
		ProjectAccessStatus(w http.ResponseWriter, req *http.Request)
		ProjectAccessRequestList(w http.ResponseWriter, req *http.Request)
		ProjectAccessRequestAction(w http.ResponseWriter, req *http.Request)
		RetrieveProjectList(w http.ResponseWriter, req *http.Request)

		InsertIntoTable(w http.ResponseWriter, req *http.Request)
		RetrieveDataTable(w http.ResponseWriter, req *http.Request)
		DeleteEntryFromTable(w http.ResponseWriter, req *http.Request)

		Stop()
	}

	controllerModel struct {
		logger   *logrus.Logger
		database database.Database
		config   *model.Config
		ctx      context.Context
	}
)

func NewApiServerModel(ctx context.Context, logger *logrus.Logger, database database.Database, config *model.Config) Controller {
	return &controllerModel{
		logger:   logger,
		database: database,
		config:   config,
		ctx:      ctx,
	}
}

func (controller *controllerModel) Response(w http.ResponseWriter, code int, payload interface{}) {
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

		if controller.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusBadRequest)
	case http.StatusUnauthorized:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusUnauthorized
		resultStatus.Message = "Unauthorized"
		if controller.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusUnauthorized)
	case http.StatusNotFound:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusNotFound
		resultStatus.Message = "Not found"
		if controller.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusNotFound)
	case http.StatusInternalServerError:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusInternalServerError
		resultStatus.Message = "Internal server error"
		if controller.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusInternalServerError)
	case http.StatusConflict:
		resultStatus.Status = common.FAILED
		resultStatus.StatusCode = http.StatusConflict
		resultStatus.Message = "Conflict"
		if controller.config.ErrorDataResponse {
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
		if controller.config.ErrorDataResponse {
			resultStatus.Data = payload
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(resultStatus)
	return
}

func (controller *controllerModel) authorization(w http.ResponseWriter, authorizationHeader, endpointName string) (int, error) {
	if authorizationHeader == "" {
		err := fmt.Errorf("missing authorization header")
		return -1, err
	}
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return -1, fmt.Errorf("invalid authorization scheme")
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
	if strings.TrimSpace(tokenString) == "" {
		return -1, fmt.Errorf("empty bearer token")
	}
	userId, err := controller.verifyToken(tokenString)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func (controller *controllerModel) verifyToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return controller.config.JWT.Secret, nil
	})
	if err != nil {
		return -1, err
	}
	claims, ok := token.Claims.(*JwtClaims)
	if !ok || !token.Valid {
		return -1, fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}

func (controller *controllerModel) createToken(userId int, name, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":    userId,
			"name":  name,
			"email": email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(controller.config.JWT.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (controller *controllerModel) checkPermission(userId int, projectId string, minRequiredPermission common.PERMISSION) (bool, int, error) {
	controller.logger.Debugf("Checking if user %d has permission to project %s", userId, projectId)
	code, err := controller.database.CheckPermission(controller.ctx, userId, projectId, minRequiredPermission)
	if err != nil {
		return false, code, err
	}
	return code == http.StatusOK, code, nil
}

func (controller *controllerModel) createTableNameFromName(displayName string, projectId int) string {
	tableName := strings.ReplaceAll(displayName, "-", "_")
	tableName = strings.ReplaceAll(tableName, " ", "_")

	return fmt.Sprintf("%v_%s", projectId, tableName)
}

func (controller *controllerModel) Stop() {
	if err := controller.database.Close(); err != nil {
		controller.logger.Errorf("db close error: %v", err)
	} else {
		controller.logger.Debugf("db closed; stopping api-server")
	}
}
