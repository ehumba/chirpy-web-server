package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

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
