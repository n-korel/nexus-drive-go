package domain

import (
	"context"

	tripTypes "github.com/n-korel/nexus-drive-go/services/trip-service/pkg/types"
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
	"github.com/n-korel/nexus-drive-go/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)



type TripModel struct {
	ID primitive.ObjectID
	UserID string
	Status string
	RideFare *RideFareModel
	Driver *pb.TripDriver
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error

	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)
}