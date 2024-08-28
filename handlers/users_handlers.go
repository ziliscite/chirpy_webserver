package handlers

import (
	"chirpy/helpers"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
)

// UsersRequestBody Login/Register request body
type UsersRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UsersResponseBody struct {
	Id          string `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (cfg *ApiConfig) validateNewUsers(r *http.Request, userRequest *UsersRequestBody) (int, error) {
	err := helpers.RequestBodyValidator(r, &userRequest)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid request body")
	}

	if !helpers.IsValidEmail(userRequest.Email) {
		return http.StatusBadRequest, fmt.Errorf("invalid email")
	}

	err = helpers.ValidatePassword(userRequest.Password)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid password: %s", err)
	}

	return http.StatusOK, nil
}

func (cfg *ApiConfig) RegisterUsersHandler(w http.ResponseWriter, r *http.Request) {
	userRequest := UsersRequestBody{}
	code, err := cfg.validateNewUsers(r, &userRequest)
	if err != nil {
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdUser, code, err := cfg.DB.CreateUsers(userRequest.Email, hashedPassword)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	responseUser := UsersResponseBody{
		Id:          createdUser.Id,
		Email:       createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed,
	}

	helpers.RespondWithJSON(w, code, responseUser)
}

func (cfg *ApiConfig) UpdateUsersHandler(w http.ResponseWriter, r *http.Request) {
	requestHeader := r.Header.Get("Authorization")
	// Takes access token btw
	tokenString := strings.TrimPrefix(requestHeader, "Bearer ")

	token, err := cfg.validateJWTToken(tokenString)
	if err != nil {
		log.Printf(err.Error())
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Invalid token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	userUpdateRequest := UsersRequestBody{}
	code, err := cfg.validateNewUsers(r, &userUpdateRequest)
	if err != nil {
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userUpdateRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, code, err := cfg.DB.UpdateUser(userId, userUpdateRequest.Email, hashedPassword)
	if err != nil {
		log.Printf("Error updating user: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	userResponse := UsersResponseBody{
		Id:    user.Id,
		Email: user.Email,
	}

	helpers.RespondWithJSON(w, code, userResponse)
}
