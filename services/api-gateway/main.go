package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n-korel/nexus-drive-go/shared/env"
	"github.com/n-korel/nexus-drive-go/shared/messaging"
)

var (
	httpAddr = env.GetString("GATEWAY_HTTP_ADDR", ":8081")
	rabbitMQURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Start API Gateway")

	mux := http.NewServeMux()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	mux.HandleFunc("POST /trip/preview", enableCORS(handleTripPreview))
	mux.HandleFunc("POST /trip/start", enableCORS(handleTripStart))
	mux.HandleFunc("/ws/drivers", func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebSocket(w, r, rabbitmq)
	} )
	mux.HandleFunc("/ws/riders", func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	})
	mux.HandleFunc("/webhook/stripe", func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	})


	server := &http.Server{
		Addr: httpAddr,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	} ()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("Error starting Server: %v", err)
		
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