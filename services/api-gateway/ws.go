package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/n-korel/nexus-drive-go/services/api-gateway/grpc_clients"
	"github.com/n-korel/nexus-drive-go/shared/contracts"
	"github.com/n-korel/nexus-drive-go/shared/proto/driver"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("No user ID provided")
		return
	}


	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Errpr reading message: %v", err)
			break
		}

		log.Printf("Received message: %s", message)
	}

}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	
	defer conn.Close()
	
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Printf("No user ID provided")
		return
	}
	
	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Printf("No package slug provided")
		return
	}
	
	ctx := r.Context()

	// Create new client for each connection
	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		driverService.Close()
		
		log.Println("Driver unregistered: ", userID)
	}()


	// CALL DRIVER SERVICE
	driverData, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}


	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Errpr reading message: %v", err)
			break
		}

		log.Printf("Received message: %s", message)
	}
	
}