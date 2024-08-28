package handlers

import (
	"chirpy/helpers"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
)

type PolkaReq struct {
	Event string `json:"event"`
	Data  struct {
		UserId string `json:"user_id"`
	} `json:"data"`
}

func (cfg *ApiConfig) PolkaHandler(w http.ResponseWriter, r *http.Request) {
	requestHeader := r.Header.Get("Authorization")
	// Takes polka token btw
	tokenString := strings.TrimPrefix(requestHeader, "ApiKey ")
	err := validatePolka(tokenString)
	if err != nil {
		log.Println(err)
		helpers.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	polkaEvent := PolkaReq{}
	err = helpers.RequestBodyValidator(r, &polkaEvent)
	if err != nil {
		log.Printf("invalid request body: %s", err)
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if polkaEvent.Event != "user.upgraded" {
		log.Printf("invalid event: %s", polkaEvent.Event)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, code, err := cfg.DB.UpgradeUser(polkaEvent.Data.UserId)
	if err != nil {
		log.Printf("failed to upgrade user: %s", err)
		helpers.RespondWithError(w, code, err.Error())
		return
	}

	userResp := UsersResponseBody{
		Id:          polkaEvent.Data.UserId,
		Email:       user.Email,
		IsChirpyRed: true,
	}

	helpers.RespondWithJSON(w, code, userResp)
}

func validatePolka(reqToken string) error {
	polkaKey := os.Getenv("POLKA_SECRET")

	if reqToken != polkaKey {
		return errors.New("invalid polka key")
	}

	return nil
}
