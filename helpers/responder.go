package helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}

	w.WriteHeader(code)
	respBody := ErrorResponse{msg}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(dat)
	if err != nil {
		log.Printf("Error writting data: %s", err)
		w.WriteHeader(500)
		return
	}

	return
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	// Handle 204 No Content response
	// I will not use this function to response with 204, but just in case I forgot, so that we can avoid error
	if code == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_, err = w.Write(dat)
	if err != nil {
		log.Printf("Error writting data: %s", err)
		w.WriteHeader(500)
		return
	}

	return
}
