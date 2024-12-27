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

type DummyRequest struct{}

var mockHandler = func(ctx context.Context, req interface{}) (interface{}, error) { return req, nil }

func TestAuthJwtInterceptor(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := CreateJWTToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})
		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("invalid audience", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := CreateJWTToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{"invalid-audience"},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("invalid issuer", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := CreateJWTToken(t,
			&CreateTokenParams{
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid issuer claim")
	})

	t.Run("invalid secret key", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := CreateJWTToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: "wrong-secret-key",
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error parsing token")
	})

	t.Run("invalid token format", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs("authorization", "Bearer bad_token")
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "token is malformed")
	})

	t.Run("missing authorization header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs()
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "missing authorization header")
	})

	t.Run("invalid authorization bearer header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs("authorization", "token invalid")
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("invalid claims format for audience", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// Create a token with invalid format for the "aud" claim (e.g., not an array)
		claims := &jwt.RegisteredClaims{
			Issuer:    os.Getenv("MIST_API_JWT_ISSUER"),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(os.Getenv("MIST_API_JWT_SECRET_KEY")))

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err = AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("missing header errors", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		_, err := AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})
}
