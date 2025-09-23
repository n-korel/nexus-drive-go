package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/grpc"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/repository"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/service"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9083"

func main() {
	inMemoryRepo := repository.NewInMemoryRepository()
	serv := service.NewService(inMemoryRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Starting gRPC server
	grpcServer := grpcserver.NewServer()

	grpc.NewGRPCHandler(grpcServer, serv)

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