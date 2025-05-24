package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/producer"
	"mist/src/rpcs"
)

func InitializeServer(kp *producer.KafkaProducer) {
	// ----- DB CONNECTION -----
	// Set up the database connection pool
	dbConn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	defer dbConn.Close()

	// Check if db connection was successful
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

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
	// Register the gRPC services
	rpcs.RegisterGrpcServices(s, dbConn, kp)

	// Start the gRPC server
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {

	// ----- KAFKA PRODUCER -----
	p, err := connectToKafka([]string{os.Getenv("KAFKA_MAIN_BROKER")})
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}

	kp := producer.NewKafkaProducer(p, os.Getenv("KAFKA_EVENT_TOPIC"))
	defer kp.Producer.Close()

	// Check if producer was able to connect to kafka server successfully
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}

	InitializeServer(kp)
}

func connectToKafka(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	return sarama.NewSyncProducer(brokers, config)
}
