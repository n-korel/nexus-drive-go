package messaging

import (
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_drivers"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}