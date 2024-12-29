package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/middleware"
	pb_channel "mist/src/protos/channel/v1"
	pb_server "mist/src/protos/server/v1"
	"mist/src/rpcs"
)

func InitializeServer() {
	dbcPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	defer dbcPool.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.AuthJwtInterceptor))

	pb_server.RegisterServerServiceServer(s, &rpcs.AppserverGRPCService{DbcPool: dbcPool})
	pb_channel.RegisterChannelServiceServer(s, &rpcs.ChannelGRPCService{DbcPool: dbcPool})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	InitializeServer()
}
