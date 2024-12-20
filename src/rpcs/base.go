package rpcs

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

type GRPCServer struct {
	pb_mistbe.UnimplementedMistBEServiceServer
	dbc_pool *pgxpool.Pool
}

func logRequestBody(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Log the metadata (headers) and the request body (payload)
	// md, ok := metadata.FromIncomingContext(ctx)
	// if ok {
	// 	log.Printf("Metadata: %v", md)
	// }

	// // Log the request body. This assumes req implements proto.Message
	// log.Printf("Request Body: %v", req)

	// Proceed with the handler
	return handler(ctx, req)
}

func ErrorHandler(err error) error {
	parsed_error := service.ParseServiceError(err.Error())
	if parsed_error == service.ValidationError {
		return status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	return status.Errorf(codes.Unknown, "%s", err.Error())
}

func InitializeGRPCServer() {
	dbc_pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(logRequestBody))
	pb_mistbe.RegisterMistBEServiceServer(s, &GRPCServer{dbc_pool: dbc_pool})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
