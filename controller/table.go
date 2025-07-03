package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
	"strconv"
	"strings"
)

func (apiServer *APIServerModel) CreateNewTable(w http.ResponseWriter, r *http.Request) {
	endpointName := "Create-New-Table"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := r.Header.Get("Authorization")
	userId, err := apiServer.authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := r.URL.Query()
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

	tableModel.TableName = apiServer.createTableNameFromName(tableModel.DisplayName, projectId)

	hasPermission, err := apiServer.checkPermission(userId, paramsProjectId, common.OWNER)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

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

func (apiServer *APIServerModel) DropTable(w http.ResponseWriter, r *http.Request) {
	endpointName := "Drop Table"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := r.Header.Get("Authorization")
	userId, err := apiServer.authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(r.Body)
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

	projectId := strings.Split(tableModel.TableName, "_")

	hasPermission, err := apiServer.checkPermission(userId, projectId[0], common.OWNER)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
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
	userId, err := apiServer.authorization(w, tokenString, endpointName)
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

	hasPermission, err := apiServer.checkPermission(userId, retrieveTableListModel.ProjectId, common.ADMIN)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
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
