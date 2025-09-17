package domain

import (
	"context"

	"github.com/n-korel/nexus-drive-go/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)



type TripModel struct {
	ID primitive.ObjectID
	UserID string
	Status string
	RideFare *RideFareModel
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.OsrmApiResponse, error)
}