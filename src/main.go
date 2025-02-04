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
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
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

	pb_appserver.RegisterServerServiceServer(s, &rpcs.AppserverGRPCService{DbcPool: dbcPool})
	pb_channel.RegisterChannelServiceServer(s, &rpcs.ChannelGRPCService{DbcPool: dbcPool})
	pb_appuser.RegisterAppuserServiceServer(s, &rpcs.AppuserGRPCService{DbcPool: dbcPool})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	InitializeServer()
}
