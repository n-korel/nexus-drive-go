package messaging

import (
	pbd "github.com/n-korel/nexus-drive-go/shared/proto/driver"
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_drivers"
	DriverCmdTripRequestQueue = "driver_cmd_trip_request"
	DriverTripResponseQueue = "driver_trip_response"
	NotifyDriverNoDriversFoundQueue = "notify_driver_no_drivers_found"
	NotifyDriverAssignQueue = "notify_driver_assign"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
}