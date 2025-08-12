package main

import (
	"encoding/json"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/ehumba/chirpy-web-server/internal/auth"
	"github.com/ehumba/chirpy-web-server/internal/database"
	"github.com/google/uuid"
)

func (a *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "could not decode email")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "invalid password")
		return
	}

	newUserDb, err := a.dbQueries.CreateUser(r.Context(), database.CreateUserParams{params.Email, hashedPassword})
	if err != nil {
		respondWithError(w, 500, "failed to create new user")
		return
	}

	newUser := User{
		ID:        newUserDb.ID,
		CreatedAt: newUserDb.CreatedAt,
		UpdatedAt: newUserDb.UpdatedAt,
		Email:     newUserDb.Email,
	}

	respondWithJSON(w, 201, newUser)
}

func (a *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "could not decode parameters")
		return
	}

	// authenticate
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authentication header")
		return
	}

	id, err := auth.ValidateJWT(bearerToken, a.secret)
	if err != nil {
		respondWithError(w, 401, "no authorization")
		return
	}

	// check if the chirp is valid
	cleansedBody := removeProfane(params.Body)

	charCount := utf8.RuneCountInString(cleansedBody)
	if charCount > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// if valid, respond.
	newChirpDb, err := a.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleansedBody, UserID: id})
	if err != nil {
		respondWithError(w, 500, "failed to create new chirp")
		return
	}
	newChirp := Chirp{
		ID:        newChirpDb.ID,
		CreatedAt: newChirpDb.CreatedAt,
		UpdatedAt: newChirpDb.UpdatedAt,
		Body:      newChirpDb.Body,
		UserID:    newChirpDb.UserID,
	}

	respondWithJSON(w, 201, &newChirp)
}

func (a *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirpsDB, err := a.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to get chirps")
		return
	}

	chirpsArray := []Chirp{}
	for _, chirpDB := range chirpsDB {
		chirp := Chirp{
			ID:        chirpDB.ID,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
			Body:      chirpDB.Body,
			UserID:    chirpDB.UserID,
		}
		chirpsArray = append(chirpsArray, chirp)
	}

	respondWithJSON(w, 200, chirpsArray)
}

func (a *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("chirpID")
	chirpDB, err := a.dbQueries.GetChirp(r.Context(), uuid.MustParse(idString))
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	chirp := Chirp{
		ID:        chirpDB.ID,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
		Body:      chirpDB.Body,
		UserID:    chirpDB.UserID,
	}
	respondWithJSON(w, 200, chirp)
}

func (a *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "could not decode parameters")
		return
	}

	userDB, err := a.dbQueries.LookUpByEmail(r.Context(), params.Email)
	hashErr := auth.CheckPasswordHash(params.Password, userDB.HashedPassword)
	if err != nil || hashErr != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	authToken, err := auth.MakeJWT(userDB.ID, a.secret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "failed to create authentication token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "failed to create refresh token")
		return
	}

	refreshTokenParams := database.GenerateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userDB.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	_, err = a.dbQueries.GenerateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, 500, "failed to save refresh token")
		return
	}

	user := User{
		ID:        userDB.ID,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
		Email:     userDB.Email,
	}

	resStruct := struct {
		User         `json:",inline"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		User:         user,
		Token:        authToken,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, 200, resStruct)
}
