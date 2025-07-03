package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
)

func (apiServer *APIServerModel) CreateNewProject(w http.ResponseWriter, req *http.Request) {
	endpointName := "Create New Project"
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

func (apiServer *APIServerModel) EditProject(w http.ResponseWriter, req *http.Request) {
	endpointName := "Edit Project"
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
	}

	projectModel := &model.ProjectModel{}
	err = json.Unmarshal(body, projectModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if projectModel.Id == "" || projectModel.Name == "" || projectModel.Desc == "" {
		err = fmt.Errorf("bad request")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectModel.UserId = userId

	statusCode, err := apiServer.database.EditProject(apiServer.ctx, projectModel)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return

	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, nil)
	return
}

func (apiServer *APIServerModel) CheckProjectPermission(w http.ResponseWriter, req *http.Request) {
	endpointName := "Check Project Permission"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := apiServer.authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	checkPermission, err := apiServer.checkPermission(userId, projectId, common.ADMIN)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	apiServer.Response(w, http.StatusOK, checkPermission)
	return

}

func (apiServer *APIServerModel) ProjectAccessRequest(w http.ResponseWriter, req *http.Request) {
	endpointName := "project access request"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := apiServer.authorization(w, tokenString, endpointName)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	err = apiServer.database.ProjectAccessRequest(apiServer.ctx, userId, projectId)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiServer.Response(w, http.StatusOK, nil)
	return

}

func (apiServer *APIServerModel) RetrieveProjectList(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Project List"
	apiServer.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := apiServer.authorization(w, tokenString, endpointName)
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
