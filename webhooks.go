package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (a *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}

	type reqParams struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "invalid request")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, 400, "invalid user id format")
		return
	}

	_, err = a.dbQueries.LookUpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "user not found")
		return
	}

	err = a.dbQueries.MakeChirpyRed(r.Context(), id)
	if err != nil {
		respondWithError(w, 400, "unable to update user to chirpy red")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
