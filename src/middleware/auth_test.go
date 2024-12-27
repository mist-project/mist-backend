package middleware

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type DummyRequest struct {
}

var mockHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func TestAuthJwtInterceptorValidToken(t *testing.T) {
	ctx := context.Background()

	tokenStr := CreateJWTToken(t,
		&CreateTokenParams{
			iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
		})
	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))

	ctx = metadata.NewIncomingContext(ctx, headers)

	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	assert.Nil(t, err)
}

func TestAuthJwtInterceptorInvalidAudience(t *testing.T) {
	ctx := context.Background()

	// Create a token with invalid audience
	tokenStr := CreateJWTToken(t,
		&CreateTokenParams{
			iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			aud:       []string{"invalid-audience"},
			secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
		})

	// Set up metadata with the invalid token
	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error contains "invalid audience"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid audience claim")
}

func TestAuthJwtInterceptorInvalidIssuer(t *testing.T) {
	ctx := context.Background()

	// Create a token with invalid issuer
	tokenStr := CreateJWTToken(t,
		&CreateTokenParams{
			aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
		})

	// Set up metadata with the invalid token
	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error contains "invalid issuer"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid issuer claim")
}

func TestAuthJwtInterceptorInvalidSecretKey(t *testing.T) {
	ctx := context.Background()

	// Create a valid token, but try decoding with a wrong secret key
	tokenStr := CreateJWTToken(t,
		&CreateTokenParams{
			iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			secretKey: "wrong-secret-key",
		})

	// Set up metadata with the invalid token
	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error contains "error parsing token"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error parsing token")
}

func TestAuthJwtInterceptorInvalidTokenFormat(t *testing.T) {
	ctx := context.Background()

	// Create metadata with an invalid token format (e.g., not a Bearer token)
	headers := metadata.Pairs("authorization", "Bearer bad_token")
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error matches "invalid token"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "token is malformed")
}

func TestAuthJwtInterceptorMissingAuthorizationHeader(t *testing.T) {
	ctx := context.Background()

	headers := metadata.Pairs()
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error matches "invalid token"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing authorization header")
}

func TestAuthJwtInterceptorInvalidAuthorizationBearerHeader(t *testing.T) {
	ctx := context.Background()

	headers := metadata.Pairs("authorization", "token invalid")
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error matches "invalid token"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

// Test case: Invalid Claims Format for `aud`
func TestAuthJwtInterceptorInvalidClaimsFormatForAudience(t *testing.T) {
	ctx := context.Background()

	// Create a token with invalid format for the "aud" claim (e.g., not an array)

	claims := &jwt.RegisteredClaims{
		Issuer:    os.Getenv("MIST_API_JWT_ISSUER"),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	// Create a new token with specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	tokenStr, err := token.SignedString([]byte(os.Getenv("MIST_API_JWT_SECRET_KEY")))

	// Set up metadata with the invalid token
	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
	ctx = metadata.NewIncomingContext(ctx, headers)

	// Call the interceptor
	_, err = AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid audience claim")
}

func TestAuthJwtInterceptorMissingHeaderErrors(t *testing.T) {
	ctx := context.Background()

	// Call the interceptor
	_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

	// Assert that the error matches "invalid token"
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unauthenticated")
}

// func TestAuthJwtInterceptorInvalidSigningMethod(t *testing.T) {
// 	ctx := context.Background()

// 	// Create a token with an invalid signing method (e.g., RS256 instead of HS256)
// 	claims := jwt.MapClaims{
// 		"aud": []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
// 		"iss": os.Getenv("MIST_API_JWT_ISSUER"),
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
// 	tokenStr, err := token.SignedString([]byte("wrong-secret-key")) // This is an invalid key for RS256
// 	assert.NoError(t, err)

// 	// Set up metadata with the token
// 	headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
// 	ctx = metadata.NewIncomingContext(ctx, headers)

// 	// Call the interceptor
// 	_, err = AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

// 	// Assert that the error contains "unexpected signing method"
// 	assert.NotNil(t, err)
// 	assert.Contains(t, err.Error(), "unexpected signing method")
// }
