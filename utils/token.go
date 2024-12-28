package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Generator interface {
	Generate(username string) (string, error)
	Validate(providedToken string) (bool, error)
}

type JwtGenerator struct {
	signingKey string
}

func NewJwtGenerator(key string) Generator {
	return &JwtGenerator{
		signingKey: key,
	}
}

func (gen *JwtGenerator) Generate(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   username,                                          // Username as the subject
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // Token expiry
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "conversion-api",
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(gen.signingKey))

	return tokenString, err
}

func (gen *JwtGenerator) Validate(providedToken string) (bool, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(gen.signingKey), nil
	}
	jwtParsed, err := jwt.ParseWithClaims(providedToken, jwt.MapClaims{}, keyFunc)
	if err != nil {

		return false, jwt.ErrTokenMalformed
	}

	_, ok := jwtParsed.Claims.(jwt.MapClaims)
	if !ok {
		return false, jwt.ErrTokenMalformed
	}

	return ok, nil
}
