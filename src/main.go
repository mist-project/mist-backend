package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/rpcs"
)

func InitializeServer() {
	dbConn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	defer dbConn.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//TODO: add these registers to its own function
	s := grpc.NewServer(rpcs.BaseInterceptors())
	rpcs.RegisterGrpcServices(s, dbConn)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	InitializeServer()
}
