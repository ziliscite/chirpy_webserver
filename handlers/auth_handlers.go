package handlers

import (
	"chirpy/database"
	"chirpy/helpers"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
)

type LoginResponseBody struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
}

func (cfg *ApiConfig) authenticateUser(r *http.Request, userRequest *UsersRequestBody) (database.User, int, error) {
	err := helpers.RequestBodyValidator(r, &userRequest)
	if err != nil {
		return database.User{}, http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err)
	}

	user, code, err := cfg.DB.GetUser(userRequest.Email)
	if err != nil {
		return database.User{}, code, fmt.Errorf("%s", err)
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(userRequest.Password))
	if err != nil {
		return database.User{}, http.StatusUnauthorized, fmt.Errorf("%s", err)
	}

	return user, http.StatusOK, nil
}

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
	userRequest := UsersRequestBody{}
	user, code, err := cfg.authenticateUser(r, &userRequest)
	if err != nil {
		log.Printf("error authenticating user: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	refreshToken, err := helpers.GenerateRefreshToken(user.Id)
	if err != nil {
		log.Printf("error generating refresh token: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	err = cfg.DB.StoreRefreshToken(refreshToken)
	if err != nil {
		log.Printf("error storing refresh token: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	// After generating the refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken.Token,
		HttpOnly: true,
		Path:     "/refresh",
		Expires:  refreshToken.ExpireAt,
	})

	accessToken, err := helpers.GenerateJWTToken(user.Id, cfg.JWTSecret)
	if err != nil {
		log.Printf("Error generating token: %s", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Should probably also return its expiredAt
	responseUser := LoginResponseBody{
		Id:           user.Id,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}

	helpers.RespondWithJSON(w, http.StatusOK, responseUser)
}

// RefreshHandler Generates a new access token using the refresh token
func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	type RespToken struct {
		AccessToken string `json:"token"`
	}

	/*	// Can be uncommented for testing purpose
		requestHeader := r.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(requestHeader, "Bearer ")
	*/

	// Extract the refresh token from the cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		log.Printf("error retrieving refresh token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	tokenString := cookie.Value

	refreshToken, err := cfg.DB.GetRefreshToken(tokenString)
	if err != nil {
		log.Printf("error getting refresh token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	accessToken, err := helpers.GenerateJWTToken(refreshToken.UserId, cfg.JWTSecret)
	if err != nil {
		log.Printf("error generating token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	responseToken := RespToken{
		AccessToken: accessToken,
	}

	helpers.RespondWithJSON(w, http.StatusOK, responseToken)
}

func (cfg *ApiConfig) RevokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	// takes the refresh token
	requestHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(requestHeader, "Bearer ")

	refreshToken, err := cfg.DB.GetRefreshToken(tokenString)
	if err != nil {
		log.Printf("error getting refresh token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	err = cfg.DB.RevokeRefreshToken(refreshToken.Token)
	if err != nil {
		log.Printf("error revoking token: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (cfg *ApiConfig) validateJWTToken(tokenString string) (*jwt.Token, error) {
	claims := &jwt.RegisteredClaims{}
	mySigningKey := []byte(cfg.JWTSecret)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(key *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if err != nil {
		log.Printf("Invalid token: %s", err)
		return nil, fmt.Errorf("invalid token")
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer != "chirpy" {
		log.Printf("Invalid auth: %s", err)
		return nil, fmt.Errorf("invalid token")
	}

	if !token.Valid {
		log.Printf("Invalid token")
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
