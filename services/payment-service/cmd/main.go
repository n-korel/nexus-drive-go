package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/n-korel/nexus-drive-go/services/payment-service/internal/events"
	"github.com/n-korel/nexus-drive-go/services/payment-service/internal/infrastructure/stripe"
	"github.com/n-korel/nexus-drive-go/services/payment-service/internal/service"
	"github.com/n-korel/nexus-drive-go/services/payment-service/pkg/types"
	"github.com/n-korel/nexus-drive-go/shared/env"
	"github.com/n-korel/nexus-drive-go/shared/messaging"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "http://localhost:3000")

	// Stripe config
	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=cancel"),
	}

	if stripeCfg.StripeSecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
		return
	}

	// Stripe processor
	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	// Service
	serv := service.NewPaymentService(paymentProcessor)

	
	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	
	log.Println("Starting RabbitMQ connection")

	// Trip Consumer
	tripConsumer := events.NewTripConsumer(rabbitmq, serv)
	go tripConsumer.Listen()
	

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down payment service...")
}