package middleware_test

import (
	"context"
	"fmt"
	"mist/src/middleware"
	"mist/src/testutil"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthJwtInterceptor(t *testing.T) {
	interceptor := middleware.AuthJwtInterceptor()

	t.Run("valid_token", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		token, _ := testutil.CreateJwtToken(t,
			&testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})
		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("invalid_audience", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		token, _ := testutil.CreateJwtToken(t,
			&testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{"invalid-audience"},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("invalid_issuer", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		token, _ := testutil.CreateJwtToken(t,
			&testutil.CreateTokenParams{
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid issuer claim")
	})

	t.Run("invalid_secret_key", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		token, _ := testutil.CreateJwtToken(t,
			&testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: "wrong-secret-key",
			})

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

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
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

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
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

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
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

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

		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token, err := tok.SignedString([]byte(os.Getenv("MIST_API_JWT_SECRET_KEY")))

		headers := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx = metadata.NewIncomingContext(ctx, headers)

		// ACT
		_, err = interceptor(ctx, dummyRequest{}, nil, MockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid audience claim")
	})

	t.Run("missing_header_errors", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		_, err := interceptor(ctx, dummyRequest{}, nil, MockHandler)

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})
}

func TestGetJWTClaims(t *testing.T) {
	t.Run("can_successfully_get_claims_from_context", func(t *testing.T) {
		// ARRANGE
		c := &middleware.CustomJWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "dummy issuer",
				Audience: jwt.ClaimStrings{"oo aud"},

				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID: uuid.NewString(),
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsK, c)

		// ACT
		ctxClaims, err := middleware.GetJWTClaims(ctx)

		// ASSERT
		assert.NotNil(t, ctxClaims)
		assert.Nil(t, err)
	})

	t.Run("invalid_claims_return_error", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsK, "boom")

		// ACT
		ctxClaims, err := middleware.GetJWTClaims(ctx)

		// ASSERT
		assert.Nil(t, ctxClaims)
		assert.NotNil(t, err)
	})
}

func TestGetUserId(t *testing.T) {

	t.Run("it_returns_the_user_id_from_context", func(t *testing.T) {
		// ARRANGE
		c := &middleware.CustomJWTClaims{
			UserID: uuid.NewString(),
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsK, c)

		// ACT
		userId := middleware.GetUserId(ctx)

		// ASSERT
		assert.NotEmpty(t, userId, "Expected a non-empty user ID")
	})

	t.Run("it_returns_NA_when_claims_are_not_present", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()

		// ACT
		userId := middleware.GetUserId(ctx)

		// ASSERT
		assert.Equal(t, "N/A", userId, "Expected 'N/A' when claims are not present")
	})

	t.Run("it_returns_NA_when_user_id_is_empty", func(t *testing.T) {
		// ARRANGE
		c := &middleware.CustomJWTClaims{
			UserID: "",
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.JwtClaimsK, c)

		// ACT
		userId := middleware.GetUserId(ctx)

		// ASSERT
		assert.Equal(t, "N/A", userId, "Expected 'N/A' when user ID is empty")
	})
}
