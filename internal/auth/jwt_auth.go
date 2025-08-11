package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentTime := jwt.NewNumericDate(time.Now())
	expTime := jwt.NewNumericDate(time.Now().Add(expiresIn))
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  currentTime,
		ExpiresAt: expTime,
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	callback := jwt.Keyfunc(func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, callback)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("error")
	}

	idString := claims.Subject
	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject id: %v", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return "", fmt.Errorf("invalid token")
	}

	return token, nil
}
