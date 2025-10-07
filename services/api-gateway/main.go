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
	"github.com/n-korel/nexus-drive-go/shared/tracing"
)

var (
	httpAddr = env.GetString("GATEWAY_HTTP_ADDR", ":8081")
	rabbitMQURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Start API Gateway")

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName: "api-gateway",
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

	mux := http.NewServeMux()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	mux.Handle("/trip/preview", tracing.WrapHandlerFunc(enableCORS(handleTripPreview), "/trip/preview"))
	mux.Handle("/trip/start", tracing.WrapHandlerFunc(enableCORS(handleTripStart), "/trip/start"))
	mux.Handle("/ws/drivers", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebSocket(w, r, rabbitmq)
	}, "/ws/drivers"))
	mux.Handle("/ws/riders", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	}, "/ws/riders"))
	mux.Handle("/webhook/stripe", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	}, "/webhook/stripe"))


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