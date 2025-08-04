package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"unicode/utf8"
)

func (a *apiConfig) postReqHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "could not decode parameters")
		return
	}

	cleansedBody := removeProfane(params.Body)

	resBody := responseBody{
		Valid:       true,
		CleanedBody: cleansedBody,
	}

	charCount := utf8.RuneCountInString(params.Body)
	if charCount > 140 {
		resBody.Valid = false
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, resBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func removeProfane(post string) string {
	forbiddenWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range forbiddenWords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		post = re.ReplaceAllString(post, "****")
	}
	return post
}

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
	}

	newUser := User{
		ID:        newUserDb.ID,
		CreatedAt: newUserDb.CreatedAt,
		UpdatedAt: newUserDb.UpdatedAt,
		Email:     newUserDb.Email,
	}

	respondWithJSON(w, 201, newUser)
}
