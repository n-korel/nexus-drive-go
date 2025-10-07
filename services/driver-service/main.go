package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/n-korel/nexus-drive-go/shared/env"
	"github.com/n-korel/nexus-drive-go/shared/messaging"
	"github.com/n-korel/nexus-drive-go/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9082"

func main() {
	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName: "driver-service",
		Environment: env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)
	
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
		
	serv := NewService()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	// Starting gRPC server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGRPCHandler(grpcServer, serv)

	consumer := NewTripConsumer(rabbitmq, serv)
	go func ()  {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("Failed to listen to the message: %v", err)
		}
	}()

	log.Printf("Starting gRPC server Driver service on port %s", listener.Addr().String())

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