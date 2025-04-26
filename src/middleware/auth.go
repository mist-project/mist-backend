package middleware

import (
	"context"
	"fmt"
	"mist/src/errors/message"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// func logRequestBody(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	// Log the metadata (headers) and the request body (payload)
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if ok {
// 		log.Printf("Metadata: %v", md)
// 	}

// 	// Log the request body. This assumes req implements proto.Message
// 	log.Printf("Request Body: %v", req)

//		// Proceed with the handler
//		return handler(ctx, req)
//	}
const JwtClaimsK string = "jwt-token"

type CustomJWTClaims struct {
	jwt.RegisteredClaims // Embed the standard registered claims

	UserID string `json:"user_id"`
}

func AuthJwtInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := headers["authorization"]
		if len(auth) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		for _, t := range auth {
			params := strings.Split(t, " ")
			if len(params) != 2 || params[0] != "Bearer" {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			claims, err := verifyJWT(params[1])
			ctx = context.WithValue(ctx, JwtClaimsK, claims)
			if err == nil {
				// Proceed with next handler
				return handler(ctx, req)
			}

			return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
		}
	}
	return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
}

func GetJWTClaims(ctx context.Context) (*CustomJWTClaims, error) {
	claims, ok := ctx.Value(JwtClaimsK).(*CustomJWTClaims)
	if !ok {
		return nil, message.UnauthenticatedError("unable to get auth claims")
	}

	return claims, nil
}

func verifyJWT(token string) (*CustomJWTClaims, error) {
	// Parse the token
	t, err := jwt.ParseWithClaims(token, &CustomJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		// TODO: we will need this in the future, for now skip
		// if token.Method != jwt.SigningMethodHS256 {
		// 	return nil, message.UnauthenticatedError("unexpected signing method: %v", token.Header["alg"])
		// }
		// Return the secret key to validate the token's signature
		return []byte(os.Getenv("MIST_API_JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, message.UnauthenticatedError(fmt.Sprintf("error parsing token: %v", err))
	}

	// Now validate the token's claims
	claims, err := verifyJWTTokenClaims(t)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("invalid audience claim")
	}

	// Validate the issuer (iss) claim
	if claims.Issuer != os.Getenv("MIST_API_JWT_ISSUER") {
		return nil, message.UnauthenticatedError("invalid issuer claim")
	}

	// AuthJWTClaims
	return claims, nil
}
