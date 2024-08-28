package handlers

import (
	"chirpy/helpers"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ChirpsRequestBody struct {
	Body string `json:"body"`
}

func (cfg *ApiConfig) PostChirpsHandler(w http.ResponseWriter, r *http.Request) {
	requestHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(requestHeader, "Bearer ")

	token, err := cfg.validateJWTToken(tokenString)
	if err != nil {
		log.Printf("Error validating JWT: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	body := ChirpsRequestBody{}
	err = helpers.RequestBodyValidator(r, &body)
	if err != nil {
		log.Printf("Invalid request body: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(body.Body) > 140 {
		log.Printf("Error request body's length exceed 140")
		helpers.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	respBody := replaceProfaneWord(body.Body)

	chirp, err := cfg.DB.CreateChirps(respBody, userId)
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}

	helpers.RespondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *ApiConfig) DeleteChirpsHandler(w http.ResponseWriter, r *http.Request) {
	requestHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(requestHeader, "Bearer ")

	token, err := cfg.validateJWTToken(tokenString)
	if err != nil {
		log.Printf("Error validating JWT: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %s", err)
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	pathValue := r.PathValue("id")
	id, err := strconv.Atoi(pathValue)
	if err != nil {
		log.Printf("Error: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Error getting chirp: "+err.Error())
		return
	}

	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Error getting chirp: "+err.Error())
		return
	}

	if chirp.UserId != userId {
		log.Printf("Error deleting chirp, unauthorized user")
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	err = cfg.DB.DeleteChirpy(chirp.Id)
	if err != nil {
		log.Printf("Error deleting chirp: %s", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error deleting chirp: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func replaceProfaneWord(s string) string {
	badWord := []string{"kerfuffle", "sharbert", "fornax"}
	dirtyS := strings.Split(strings.ToLower(s), " ")

	for _, b := range badWord {
		for i, d := range dirtyS {
			if strings.Contains(d, b) {
				dirtyS[i] = strings.Replace(dirtyS[i], d, "****", -1)
			}
		}
	}

	return strings.Join(dirtyS, " ")
}

func (cfg *ApiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	method := r.URL.Query().Get("sort")

	if authorId != "" {
		chirps, err := cfg.DB.GetChirpByAuthor(authorId, method)
		if err != nil {
			log.Printf("Error getting chirp: %s", err)
			helpers.RespondWithError(w, http.StatusInternalServerError, "Error getting chirp: "+err.Error())
			return
		}

		helpers.RespondWithJSON(w, http.StatusOK, chirps)
		return
	}

	chirps, err := cfg.DB.GetChirps(method)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error getting chirp: "+err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *ApiConfig) GetChirpHandler(w http.ResponseWriter, r *http.Request) {
	pathValue := r.PathValue("id")
	id, err := strconv.Atoi(pathValue)
	if err != nil {
		log.Printf("Error: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Error getting chirp: "+err.Error())
		return
	}

	chirp, err := cfg.DB.GetChirp(id)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "Error getting chirp: "+err.Error())
		return
	}

	helpers.RespondWithJSON(w, http.StatusOK, chirp)
}
