package controller

import (
	"encoding/json"
	"fmt"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
	"strconv"
	"strings"
)

func (controller *controllerModel) CreateNewTable(w http.ResponseWriter, r *http.Request) {
	endpointName := "Create-New-Table"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := r.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var tableModel *model.CreateTableModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	params := r.URL.Query()
	paramsProjectId := params.Get("project_id")

	projectId, err := strconv.Atoi(paramsProjectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	tableModel.TableName = controller.createTableNameFromName(tableModel.DisplayName, projectId)

	hasPermission, httpCode, err := controller.checkPermission(userId, paramsProjectId, common.OWNER) // TODO: do we have to give permission to ADMIN?
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, httpCode, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	statusCode, err := controller.database.CreateTable(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, statusCode, err.Error())
		return
	}

	// TODO: do it as a transaction with above
	statusCode, err = controller.database.InsertIntoMetaData(controller.ctx, tableModel, projectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, statusCode, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusCreated, nil)

	return
}

func (controller *controllerModel) DropTable(w http.ResponseWriter, r *http.Request) {
	endpointName := "Drop Table"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := r.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var tableModel *model.DropTableModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectId := strings.Split(tableModel.TableName, "_")

	hasPermission, httpCode, err := controller.checkPermission(userId, projectId[0], common.OWNER)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, httpCode, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = controller.database.DropTable(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, nil)
	return
}

func (controller *controllerModel) RetrieveTableListFromMetaData(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Table List"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var retrieveTableListModel *model.RetrieveTableListModel

	if err := dec.Decode(&retrieveTableListModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if retrieveTableListModel.ProjectId == "" {
		err := fmt.Errorf("project ID cannot be empty")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	hasPermission, httpCode, err := controller.checkPermission(userId, retrieveTableListModel.ProjectId, common.ADMIN)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, httpCode, err.Error())
		return
	}

	if !hasPermission {
		err = fmt.Errorf("user does not have permission to access this project")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	tableList, statusCode, err := controller.database.RetrieveTableListFromMetaData(controller.ctx, retrieveTableListModel.ProjectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, statusCode, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, tableList)
	return
}
