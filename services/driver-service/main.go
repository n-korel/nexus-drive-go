package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9082"

func main() {

	serv := NewService()

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
	NewGRPCHandler(grpcServer, serv)

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