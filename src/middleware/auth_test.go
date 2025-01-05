package middleware_test

import (
	"context"
	"fmt"
	"mist/src/middleware"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type DummyRequest struct{}

var mockHandler = func(ctx context.Context, req interface{}) (interface{}, error) { return req, nil }

func TestAuthJwtInterceptor(t *testing.T) {
	t.Run("valid_token", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := createJwtToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})
		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("invalid_audience", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := createJwtToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{"invalid-audience"},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("invalid_issuer", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := createJwtToken(t,
			&CreateTokenParams{
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid issuer claim")
	})

	t.Run("invalid_secret_key", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		tokenStr := createJwtToken(t,
			&CreateTokenParams{
				iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				secretKey: "wrong-secret-key",
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tokenStr))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "error parsing token")
	})

	t.Run("invalid_token_format", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs("authorization", "Bearer bad_token")
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "token is malformed")
	})

	t.Run("missing_authorization_header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs()
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "missing authorization header")
	})

	t.Run("invalid_authorization_bearer_header", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		headers := metadata.Pairs("authorization", "token invalid")
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("invalid_claims_format_for_audience", func(t *testing.T) {
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
		_, err = middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("missing_header_errors", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		_, err := middleware.AuthJwtInterceptor(ctx, DummyRequest{}, nil, mockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})
}

func TestGetJWTClaims(t *testing.T) {
	t.Run("can_successfully_get_claims_from_context", func(t *testing.T) {
		// ARRANGE
		claims := &middleware.CustomJWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "dummy issuer",
				Audience: jwt.ClaimStrings{"oo aud"},

				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID: uuid.NewString(),
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsContextKey, claims)

		// ACT
		ctxClaims, err := middleware.GetJWTClaims(ctx)

		// ASSERT
		assert.NotNil(t, ctxClaims)
		assert.Nil(t, err)
	})

	t.Run("invalid_claims_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsContextKey, "boom")

		// ACT
		ctxClaims, err := middleware.GetJWTClaims(ctx)

		// ASSERT
		assert.Nil(t, ctxClaims)
		assert.NotNil(t, err)
	})

}
