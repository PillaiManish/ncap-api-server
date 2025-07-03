package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
	"strings"
)

func (apiServer *APIServerModel) InsertIntoTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Insert Into Table"
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

	tableModel := &model.InsertIntoTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectId := strings.Split(tableModel.TableName, "_")
	hasPermission, err := apiServer.checkPermission(userId, projectId[0], common.ADMIN)
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

	err = apiServer.database.InsertIntoTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
	return
}

func (apiServer *APIServerModel) RetrieveDataTable(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Data Table List"
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

	tableModel := &model.RetrieveDataFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	hasPermission, err := apiServer.checkPermission(userId, tableModel.ProjectId, common.ADMIN)
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

	columns, data, err := apiServer.database.RetrieveDataFromTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
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

	tableModel := &model.DeleteEntryFromTableModel{}
	err = json.Unmarshal(body, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	hasPermission, err := apiServer.checkPermission(userId, tableModel.ProjectId, common.ADMIN)
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

	err = apiServer.database.DeleteEntryFromTable(apiServer.ctx, tableModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
	return
}
