package controller

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"ncap-api-server/model"
	"net/http"
)

func (controller *controllerModel) Login(w http.ResponseWriter, r *http.Request) {
	endpointName := "Login"
	controller.logger.Debugf("Received %s request", endpointName)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var checkUser *model.UserModel

	if err := dec.Decode(&checkUser); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		err := fmt.Errorf("email/password cannot be empty")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	userModel, statusCode, err := controller.database.LoginUser(controller.ctx, checkUser)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	token, err := controller.createToken(userModel.Id, userModel.Name, userModel.Email)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, statusCode, token)
	return
}

func (controller *controllerModel) Register(w http.ResponseWriter, r *http.Request) {
	endpointName := "Register"
	controller.logger.Debugf("Received %s request", endpointName)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var newUser *model.UserModel

	if err := dec.Decode(&newUser); err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		err := fmt.Errorf("name/password/email cannot be empty")
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser.Password = string(encryptPasswd)

	userId, _, err := controller.database.RegisterUser(controller.ctx, newUser)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		httpCode, message := controller.handleDatabaseError(err)
		controller.Response(w, httpCode, message)
		return
	}

	newUser.Id = int(userId)

	token, err := controller.createToken(newUser.Id, newUser.Name, newUser.Email)
	if err != nil {
		controller.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		controller.Response(w, http.StatusInternalServerError, err.Error())
		return
	}
	controller.logger.Debugf("%s: successfully processed", endpointName)
	controller.Response(w, http.StatusCreated, token)
	return
}
