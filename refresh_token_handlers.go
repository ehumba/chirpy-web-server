package main

import (
	"net/http"
	"time"

	"github.com/ehumba/chirpy-web-server/internal/auth"
)

func (a *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "no authorization header")
		return
	}

	refreshTokenDB, err := a.dbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 401, "no valid refresh token")
		return
	}

	authToken, err := auth.MakeJWT(refreshTokenDB.UserID, a.secret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "unable to create authentication token")
		return
	}

	resStruct := struct {
		Token string `json:"token"`
	}{
		Token: authToken,
	}

	respondWithJSON(w, 200, resStruct)
}

func (a *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "no authorization header")
		return
	}

	err = a.dbQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 500, "failed to revoke refresh token")
		return
	}

	w.WriteHeader(204)
}
