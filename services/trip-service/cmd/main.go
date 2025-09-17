package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	h "github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/http"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/repository"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/service"
)

func main() {


	inMemoryRepo := repository.NewInMemoryRepository()

	serv := service.NewService(inMemoryRepo)
	mux := http.NewServeMux()

	httphandler := h.HttpHandler{Service: serv}


	mux.HandleFunc("POST /preview", httphandler.HandleTripPreview)

	server := &http.Server{
		Addr: ":8083",
		Handler: mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop Server Gracefully: %v", err)
			server.Close()
		}
	}
}