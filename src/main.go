package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/rpcs"
)

func InitializeServer() {
	DbConn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	defer DbConn.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to create protovalidate validator")
	}

	//TODO: add these registers to its own function
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		middleware.AuthJwtInterceptor,
		protovalidate_middleware.UnaryServerInterceptor(validator),
	))

	pb_appserver.RegisterAppserverServiceServer(s, &rpcs.AppserverGRPCService{DbConn: DbConn})
	pb_channel.RegisterChannelServiceServer(s, &rpcs.ChannelGRPCService{DbConn: DbConn})
	pb_appuser.RegisterAppuserServiceServer(s, &rpcs.AppuserGRPCService{DbConn: DbConn})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	InitializeServer()
}
