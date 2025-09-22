package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/n-korel/nexus-drive-go/services/api-gateway/grpc_clients"
	"github.com/n-korel/nexus-drive-go/shared/contracts"
)


func handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	
	defer r.Body.Close()

	// Create new client for each connection
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()


	// CALL TRIP SERVICE
	trip, err := tripService.Client.CreateTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("Failed to start trip: %v", err)
		http.Error(w, "Failed to start trip", http.StatusInternalServerError)
		return
	}


	response := contracts.APIResponse{Data: trip}

	writeJSON(w, http.StatusCreated, response)
}


func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	
	defer r.Body.Close()
	
	// VALIDATION
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return

	}

	// Create new client for each connection
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}


	defer tripService.Close()


	// CALL TRIP SERVICE
	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("Failed to preview trip: %v", err)
		http.Error(w, "Failed to preview trip", http.StatusInternalServerError)
		return
	}


	response := contracts.APIResponse{Data: tripPreview}

	writeJSON(w, http.StatusCreated, response)
}