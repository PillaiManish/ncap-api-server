package controller

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"ncap-api-server/model"
	"net/http"
)

func (apiServer *APIServerModel) Login(w http.ResponseWriter, r *http.Request) {
	endpointName := "Login"
	apiServer.logger.Debugf("Received %s request", endpointName)

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var checkUser *model.UserModel

	if err := dec.Decode(&checkUser); err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if checkUser.Password == "" || checkUser.Email == "" {
		err := fmt.Errorf("email/password cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	userModel, statusCode, err := apiServer.database.LoginUser(apiServer.ctx, checkUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	token, err := apiServer.createToken(userModel.Id, userModel.Name, userModel.Email)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusOK, token)
	return
}

func (apiServer *APIServerModel) Register(w http.ResponseWriter, r *http.Request) {
	endpointName := "Register"
	apiServer.logger.Debugf("Received %s request", endpointName)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var newUser *model.UserModel

	if err := dec.Decode(&newUser); err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	if newUser.Name == "" || newUser.Password == "" || newUser.Email == "" {
		err := fmt.Errorf("name/password/email cannot be empty")
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusBadRequest, err.Error())
		return
	}

	encryptPasswd, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}

	newUser.Password = string(encryptPasswd)

	userId, statusCode, err := apiServer.database.RegisterUser(apiServer.ctx, newUser)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, statusCode, err.Error())
		return
	}

	newUser.Id = int(userId)

	token, err := apiServer.createToken(newUser.Id, newUser.Name, newUser.Email)
	if err != nil {
		apiServer.logger.Errorf("%s: failed due to err: %s", endpointName, err.Error())
		apiServer.Response(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiServer.logger.Debugf("%s: successfully processed", endpointName)
	apiServer.Response(w, http.StatusCreated, token)
	return
}
