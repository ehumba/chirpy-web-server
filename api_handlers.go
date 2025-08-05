package main

import (
	"encoding/json"
	"net/http"
	"unicode/utf8"

	"github.com/ehumba/chirpy-web-server/internal/database"
	"github.com/google/uuid"
)

func (a *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type reqParams struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "could not decode email")
		return
	}

	newUserDb, err := a.dbQueries.CreateUser(r.Context(), params.Email)
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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "could not decode parameters")
		return
	}

	// check if the chirp is valid
	cleansedBody := removeProfane(params.Body)

	charCount := utf8.RuneCountInString(params.Body)
	if charCount > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// if valid, respond.
	newChirpDb, err := a.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleansedBody, UserID: params.UserID})
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
