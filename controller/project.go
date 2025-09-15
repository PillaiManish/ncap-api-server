package controller

import (
	"encoding/json"
	"fmt"
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
)

func (controller *controllerModel) CreateNewProject(w http.ResponseWriter, req *http.Request) {
	endpointName := "Create New Project"
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
	var project *model.ProjectModel

	if err := dec.Decode(&project); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	project.UserId = userId

	statusCode, err := controller.database.CreateNewProject(controller.ctx, project, controller.config.ProjectConfig.Limit)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, nil)
	return
}

func (controller *controllerModel) EditProject(w http.ResponseWriter, req *http.Request) {
	endpointName := "Edit Project"
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
	var projectModel *model.ProjectModel

	if err := dec.Decode(&projectModel); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if projectModel.Id == "" || projectModel.Name == "" || projectModel.Desc == "" {
		err = fmt.Errorf("bad request")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	projectModel.UserId = userId

	statusCode, err := controller.database.EditProject(controller.ctx, projectModel)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, nil)
	return
}

func (controller *controllerModel) CheckProjectPermission(w http.ResponseWriter, req *http.Request) {
	endpointName := "Check Project Role"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	if projectId == "" {
		controller.Response(w, http.StatusBadRequest, "missing projectId")
		return
	}

	checkPermission, httpCode, err := controller.checkPermission(userId, projectId, common.ADMIN)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, httpCode, err.Error())
		return
	}

	controller.Response(w, http.StatusOK, checkPermission)
	return
}

func (controller *controllerModel) ProjectAccessRequest(w http.ResponseWriter, req *http.Request) {
	endpointName := "project access request"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	statusCode, err := controller.database.ProjectAccessRequest(controller.ctx, userId, projectId, common.ADMIN) //TODO: replace admin with dynamic
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.Response(w, statusCode, nil)
	return
}

func (controller *controllerModel) ProjectAccessStatus(w http.ResponseWriter, req *http.Request) {
	endpointName := "project access request status"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	projectExists, statusCode, err := controller.database.CheckProjectExists(controller.ctx, projectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	if !projectExists {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, "project does not exist")
		controller.Response(w, statusCode, "project does not exist")
		return
	}

	accessStatus, statusCode, err := controller.database.ProjectAccessStatus(controller.ctx, userId, projectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.Response(w, statusCode, accessStatus.String())
	return
}

func (controller *controllerModel) ProjectAccessRequestList(w http.ResponseWriter, req *http.Request) {
	endpointName := "Project-Access-Request-List"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	hasPermission, httpCode, err := controller.checkPermission(userId, projectId, common.OWNER)
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

	projectList, statusCode, err := controller.database.ProjectAccessList(controller.ctx, userId, projectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, projectList)
}

func (controller *controllerModel) ProjectAccessRequestAction(w http.ResponseWriter, req *http.Request) {
	endpointName := "Project-Access-Request-Action"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	params := req.URL.Query()
	projectId := params.Get("projectId")

	hasPermission, httpCode, err := controller.checkPermission(userId, projectId, common.OWNER)
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

	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var projectAccessAction *model.ProjectAccessRequestModel

	if err := dec.Decode(&projectAccessAction); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	statusCode, err := controller.database.ProjectAccessAction(controller.ctx, projectAccessAction, projectId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, nil)
}

func (controller *controllerModel) RetrieveProjectList(w http.ResponseWriter, req *http.Request) {
	endpointName := "Retrieve Project List"
	controller.logger.Debugf("Received %s request", endpointName)

	tokenString := req.Header.Get("Authorization")
	userId, err := controller.authorization(w, tokenString, endpointName)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusUnauthorized, err.Error())
		return
	}

	projects, statusCode, err := controller.database.RetrieveProjectList(controller.ctx, userId)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, projects)
	return
}
