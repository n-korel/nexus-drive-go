package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/events"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/grpc"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/repository"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/service"
	"github.com/n-korel/nexus-drive-go/shared/db"
	"github.com/n-korel/nexus-drive-go/shared/env"
	"github.com/n-korel/nexus-drive-go/shared/messaging"
	"github.com/n-korel/nexus-drive-go/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {
	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "trip-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	// Initialize MongoDB
	mongoClient, err := db.NewMongoClient(ctx, db.NewMongoDefaultConfig())
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB, err: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	mongoDb := db.GetDatabase(mongoClient, db.NewMongoDefaultConfig())


	// inMemoryRepo := repository.NewInMemoryRepository()
	mongoDBRepo := repository.NewMongoRepository(mongoDb)
	serv := service.NewService(mongoDBRepo)

	// Setup graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	listener, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	publisher := events.NewTripEventPublisher(rabbitmq)

	// Start driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, serv)
	go driverConsumer.Listen()

	// Start payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, serv)
	go paymentConsumer.Listen()

	// Starting gRPC server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGRPCHandler(grpcServer, serv, publisher)

	log.Printf("Starting gRPC server Trip service on port %s", listener.Addr().String())
	
	go func ()  {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("failed to serve: %v", err)
		}
	}()


	// wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down the server....")
	grpcServer.GracefulStop()
}