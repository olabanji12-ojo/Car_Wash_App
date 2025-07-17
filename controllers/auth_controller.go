package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/olabanji12-ojo/CarWashApp/models"
	"github.com/olabanji12-ojo/CarWashApp/services"
	"github.com/olabanji12-ojo/CarWashApp/utils"


	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)



// REGISTER HANDLER 

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input models.User

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logrus.Warn("Invalid register input: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// 🔒 Validate input before passing to service
	if err := input.Validate(); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call the service to handle registration
	newUser, err := services.RegisterUser(input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, newUser)
}



// LOGIN HANDLER


func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode JSON login input
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logrus.Warn("Invalid login input: ", err)
		utils.Error(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	//  Validate login input
	err := validation.ValidateStruct(&credentials,
		validation.Field(&credentials.Email, validation.Required, is.Email),
		validation.Field(&credentials.Password, validation.Required, validation.Length(6, 100)),
	)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call service to login
	token, user, err := services.LoginUser(credentials.Email, credentials.Password)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	utils.JSON(w, http.StatusOK, response)
}






