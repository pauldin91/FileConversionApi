package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer string = "conversion_api"
)

type Generator interface {
	Generate(username, role string) (string, error)
	Validate(providedToken string) (*CustomClaims, error)
}

type JwtGenerator struct {
	signingKey string
}

type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewClaims(username string, role string) *CustomClaims {

	payload := &CustomClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   issuer + "_user",
			Issuer:    issuer,
		},
	}
	return payload
}

func NewJwtGenerator(key string) Generator {
	return &JwtGenerator{
		signingKey: key,
	}
}

func (gen *JwtGenerator) Generate(username string, role string) (string, error) {
	payload := NewClaims(username, role)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(gen.signingKey))

	return tokenString, err
}

func (gen *JwtGenerator) Validate(providedToken string) (*CustomClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(gen.signingKey), nil
	}
	jwtParsed, err := jwt.ParseWithClaims(providedToken, &CustomClaims{}, keyFunc)
	if err != nil {

		return nil, jwt.ErrTokenMalformed
	}

	claims, ok := jwtParsed.Claims.(*CustomClaims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	return claims, nil
}
