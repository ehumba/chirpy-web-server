package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

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
