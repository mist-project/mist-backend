package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"mist/src/logging/logger"
	"mist/src/producer"
	"mist/src/producer/mist_redis"
	"mist/src/psql_db/db"
	"mist/src/rpcs"
)

func InitializeServer(redisClient *redis.Client) {
	// ----- DB CONNECTION -----
	// Set up the database connection pool
	dbConn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	// Check if db connection was successful
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer dbConn.Close()

	// ----- GRPC SERVER -----
	// Create a TCP listener on the specified port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("APP_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Setup the gRPC server interceptors
	interceptors, err := rpcs.BaseInterceptors()

	if err != nil {
		log.Fatalf("failed to start interceptors: %v", err)
	}

	// Create a new gRPC server with the interceptors
	s := grpc.NewServer(interceptors)

	// Setup worker pool for message production
	p := producer.NewMProducerOptions(redisClient, &producer.MProducerOptions{
		Workers:     4,
		ChannelSize: 100,
	})

	p.Wp.StartWorkers() // Start the worker pool
	defer p.Wp.Stop()

	// Register the gRPC services
	rpcs.RegisterGrpcServices(s, &rpcs.GrpcDependencies{
		Db:        db.NewQuerier(dbConn),
		MProducer: p,
	})

	// Start the gRPC server
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {

	// // ----- REDIS CLIENT PRODUCER -----
	redisClient := connectToRedis()

	logger.InitializeLogger()
	InitializeServer(redisClient)
}

func connectToRedis() *redis.Client {
	var client *redis.Client
	ctx := context.Background()

	for client == nil {
		logger.Debug("Initializing Redis client.", "SERVICE", "REDIS")
		client = mist_redis.ConnectToRedis(os.Getenv("REDIS_DB"))

		// Perform a health check by setting a key
		result, err := client.Set(ctx, "health", "check", 0).Result()

		if err != nil {
			logger.Error(fmt.Sprintf("Failed to connect to Redis: %v", err), "SERVICE", "REDIS")
			logger.Debug("Retrying in 5 seconds...", "SERVICE", "REDIS")
			client.Close()
			client = nil // Reset client to retry connection
			// Wait for 5 seconds before retrying
			<-time.After(5 * time.Second)
		}

		if result == "OK" {
			logger.Debug("Redis connection established", "SERVICE", "REDIS")
		}

		// Clean up the health check key after setting it
		client.Del(ctx, "health").Result()

	}

	return client
}
