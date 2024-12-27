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

// 	// Proceed with the handler
// 	return handler(ctx, req)
// }

func verifyJWTTokenClaims(token *jwt.Token) error {
	// Now validate the token's claims
	claims, _ := token.Claims.(jwt.MapClaims)

	// Validate aud
	validAudience := false
	auds, ok := claims["aud"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid audience claim")
	}

	// If "aud" is an array of strings, cast each element to string
	for _, aud := range auds {
		audStr, _ := aud.(string)

		if audStr == os.Getenv("MIST_API_JWT_AUDIENCE") {
			validAudience = true
			break
		}
	}

	if !validAudience {
		return fmt.Errorf("invalid audience claim")
	}

	// Validate the issuer (iss) claim
	if iss, ok := claims["iss"].(string); !ok || iss != os.Getenv("MIST_API_JWT_ISSUER") {
		return fmt.Errorf("invalid issuer claim")
	}

	return nil
}

func VerifyJWT(tokenStr string) (*jwt.Token, error) {
	// Parse the token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
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
	err = verifyJWTTokenClaims(token)
	if err != nil {
		return nil, err
	}

	return token, nil
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

			// We will be using token returned by here eventually. will add to context
			_, err := VerifyJWT(parameters[1])
			if err == nil {
				// Proceed with next handler
				return handler(ctx, req)
			}

			return nil, status.Errorf(codes.Unauthenticated, "%s", err.Error())
		}
	}
	return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
}
