package controller

import (
	"encoding/json"
	"fmt"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
	"strings"
)

func (controller *controllerModel) InsertIntoTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Insert Into Table"
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
	var tableModel *model.InsertIntoTableModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectId := strings.Split(tableModel.TableName, "_")
	hasPermission, httpCode, err := controller.checkPermission(userId, projectId[0], common.ADMIN)
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

	err = controller.database.InsertIntoTable(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, nil)
	return
}

func (controller *controllerModel) RetrieveDataTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Data Table List"
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
	var tableModel *model.RetrieveDataFromTableModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	hasPermission, httpCode, err := controller.checkPermission(userId, tableModel.ProjectId, common.ADMIN)
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

	columns, data, primaryKeys, foreignKeys, err := controller.database.RetrieveDataFromTable(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	resultData := map[string]interface{}{
		"columns":     columns,
		"rows":        data,
		"primaryKeys": primaryKeys,
		"foreignKeys": foreignKeys,
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, resultData)
	return
}

func (controller *controllerModel) DeleteEntryFromTable(w http.ResponseWriter, req *http.Request) {
	endpointName := " Delete Entry"
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
	var tableModel *model.DeleteEntryFromTableModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	hasPermission, httpCode, err := controller.checkPermission(userId, tableModel.ProjectId, common.ADMIN)
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

	err = controller.database.DeleteEntryFromTable(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, nil)
	return
}

func (controller *controllerModel) UpdateTableData(w http.ResponseWriter, req *http.Request) {
	endpointName := "Update Table Data"
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
	var tableModel *model.UpdateTableDataModel

	if err := dec.Decode(&tableModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectId := strings.Split(tableModel.TableName, "_")
	hasPermission, httpCode, err := controller.checkPermission(userId, projectId[0], common.ADMIN)
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

	// TODO: add validation

	err = controller.database.UpdateTableData(controller.ctx, tableModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusOK, nil)
	return
}
