package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CreateTokenParams struct {
	iss       string
	aud       []string
	secretKey string
}

func CreateJWTToken(t *testing.T, params *CreateTokenParams) string {
	// Define JWT claims
	claims := &jwt.RegisteredClaims{
		Issuer:   params.iss,
		Audience: params.aud,

		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	// Create a new token with specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	tokenString, err := token.SignedString([]byte(params.secretKey))
	if err != nil {
		t.Fatalf("error signing the token %v", err)
	}
	return tokenString
}
