package main

import "github.com/n-korel/nexus-drive-go/shared/types"

type previewTripRequest struct {
	UserID string `json:"userID"`
	Pickup types.Coordinate	`json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}