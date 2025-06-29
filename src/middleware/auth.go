package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/faults"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const JwtClaimsK string = "jwt-token"

type CustomJWTClaims struct {
	jwt.RegisteredClaims // Embed the standard registered claims

	UserID string `json:"user_id"`
}

func AuthJwtInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		headers, ok := metadata.FromIncomingContext(ctx)
		if ok {
			auth := headers["authorization"]
			if len(auth) == 0 {
				return nil, faults.AuthenticationError("unable to get auth claims", slog.LevelDebug)
			}

			for _, t := range auth {
				params := strings.Split(t, " ")
				if len(params) != 2 || params[0] != "Bearer" {
					return nil, faults.AuthenticationError("invalid token", slog.LevelDebug)
				}

				claims, err := verifyJWT(params[1])
				ctx = context.WithValue(ctx, JwtClaimsK, claims)
				if err == nil {
					// Proceed with next handler
					return handler(ctx, req)
				}

				return nil, faults.ExtendError(err)
			}
		}
		return nil, faults.AuthenticationError("missing or invalid authorization header", slog.LevelDebug)
	}
}

func GetJWTClaims(ctx context.Context) (*CustomJWTClaims, error) {
	claims, ok := ctx.Value(JwtClaimsK).(*CustomJWTClaims)
	if !ok {
		return nil, faults.AuthenticationError("unable to get auth claims", slog.LevelInfo)
	}

	return claims, nil
}

func GetUserId(ctx context.Context) string {
	claims, err := GetJWTClaims(ctx)

	if err != nil {
		return "N/A"
	}

	if claims.UserID == "" {
		return "N/A"
	}

	return claims.UserID
}

func verifyJWT(token string) (*CustomJWTClaims, error) {
	// Parse the token
	t, err := jwt.ParseWithClaims(token, &CustomJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		// TODO: we will need this in the future, for now skip
		// if token.Method != jwt.SigningMethodHS256 {
		// 	return nil, faults.AuthenticationError("unexpected signing method: %v", token.Header["alg"])
		// }
		// Return the secret key to validate the token's signature
		return []byte(os.Getenv("MIST_API_JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, faults.AuthenticationError(fmt.Sprintf("error parsing token: %v", err), slog.LevelInfo)
	}

	// Now validate the token's claims
	claims, err := verifyJWTTokenClaims(t)
	if err != nil {
		return nil, faults.ExtendError(err)
	}

	return claims, nil
}

func verifyJWTTokenClaims(t *jwt.Token) (*CustomJWTClaims, error) {
	// Now validate the token's claims
	claims, _ := t.Claims.(*CustomJWTClaims)

	// Validate aud
	vAud := false
	auds := claims.Audience

	// If "aud" is an array of strings, cast each element to string
	for _, aud := range auds {
		if aud == os.Getenv("MIST_API_JWT_AUDIENCE") {
			vAud = true
			break
		}
	}

	if !vAud {
		return nil, faults.AuthenticationError("invalid audience claim", slog.LevelInfo)
	}

	// Validate the issuer (iss) claim
	if claims.Issuer != os.Getenv("MIST_API_JWT_ISSUER") {
		return nil, faults.AuthenticationError("invalid issuer claim", slog.LevelInfo)
	}

	// AuthJWTClaims
	return claims, nil
}
