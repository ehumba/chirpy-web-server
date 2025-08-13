package main

import (
	"encoding/json"
	"net/http"

	"github.com/ehumba/chirpy-web-server/internal/auth"
	"github.com/ehumba/chirpy-web-server/internal/database"
	"github.com/google/uuid"
)

func (a *apiConfig) handlerUpdate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "no authorization header")
		return
	}

	userID, err := auth.ValidateJWT(token, a.secret)
	if err != nil {
		respondWithError(w, 401, "invalid user")
		return
	}

	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "could not decode update parameters")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(w, 400, "email and password are required")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "invalid password")
		return
	}

	updateParams := database.UpdateUserDataParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	err = a.dbQueries.UpdateUserData(r.Context(), updateParams)
	if err != nil {
		respondWithError(w, 500, "error while updating user data")
		return
	}

	updatedUserDb, err := a.dbQueries.LookUpByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, 500, "error while retrieving updated user data")
		return
	}

	updatedUser := User{
		ID:        updatedUserDb.ID,
		CreatedAt: updatedUserDb.CreatedAt,
		UpdatedAt: updatedUserDb.UpdatedAt,
		Email:     updatedUserDb.Email,
	}

	respondWithJSON(w, 200, updatedUser)
}

func (a *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(idString)
	if err != nil {
		respondWithError(w, 400, "invalid chirp ID format")
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authorization")
		return
	}

	userID, err := auth.ValidateJWT(authToken, a.secret)
	if err != nil {
		respondWithError(w, 401, "unauthorized access")
		return
	}

	chirpToDelete, err := a.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	if userID != chirpToDelete.UserID {
		respondWithError(w, 403, "forbidden: you can only delete your own chirps")
		return
	}

	err = a.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 500, "failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
