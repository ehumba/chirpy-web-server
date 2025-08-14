package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization")
	}

	if !strings.HasPrefix(strings.ToLower(authHeader), "apikey ") {
		return "", fmt.Errorf("invalid authorization header")
	}

	key := strings.TrimSpace(strings.TrimPrefix(authHeader, "ApiKey "))
	if key == "" {
		return "", fmt.Errorf("invalid API key")
	}

	return key, nil
}
