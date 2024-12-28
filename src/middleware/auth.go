package middleware

import (
	"context"
	"fmt"
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
const JwtClaimsContextKey string = "jwt-token"

type CustomJWTClaims struct {
	jwt.RegisteredClaims // Embed the standard registered claims

	UserID string `json:"user_id"`
}

func verifyJWTTokenClaims(token *jwt.Token) (*CustomJWTClaims, error) {
	// Now validate the token's claims
	claims, _ := token.Claims.(*CustomJWTClaims)

	// Validate aud
	validAudience := false
	auds := claims.Audience

	// If "aud" is an array of strings, cast each element to string
	for _, aud := range auds {
		if aud == os.Getenv("MIST_API_JWT_AUDIENCE") {
			validAudience = true
			break
		}
	}

	if !validAudience {
		return nil, fmt.Errorf("invalid audience claim")
	}

	// Validate the issuer (iss) claim
	if claims.Issuer != os.Getenv("MIST_API_JWT_ISSUER") {
		return nil, fmt.Errorf("invalid issuer claim")
	}

	// AuthJWTClaims
	return claims, nil
}

func VerifyJWT(tokenStr string) (*CustomJWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenStr, &CustomJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// TODO: we will need this in the future, for now skip
		// if token.Method != jwt.SigningMethodHS256 {
		// 	return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		// }
		// Return the secret key to validate the token's signature
		return []byte(os.Getenv("MIST_API_JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	// Now validate the token's claims
	claims, err := verifyJWTTokenClaims(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func AuthJwtInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if ok {
		authorization := headers["authorization"]
		if len(authorization) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		for _, token := range authorization {
			parameters := strings.Split(token, " ")
			if len(parameters) != 2 || parameters[0] != "Bearer" {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			claims, err := VerifyJWT(parameters[1])
			ctx = context.WithValue(ctx, JwtClaimsContextKey, claims)
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
	claims, ok := ctx.Value(JwtClaimsContextKey).(*CustomJWTClaims)
	if !ok {
		return nil, fmt.Errorf("unable to get auth claims")
	}

	return claims, nil
}
