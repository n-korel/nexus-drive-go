package repository

import (
	"context"
	"fmt"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/domain"

	pbd "github.com/n-korel/nexus-drive-go/shared/proto/driver"
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
)

type inMemoryRepository struct {
	trips map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		trips: make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inMemoryRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	trip, ok := r.trips[id]
	if !ok {
		return nil, nil
	}
	return trip, nil
}

func (r *inMemoryRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error {
	trip, ok := r.trips[tripID]
	if !ok {
		return fmt.Errorf("trip not found with ID: %s", tripID)
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			CarPlate:       driver.CarPlate,
			ProfilePicture: driver.ProfilePicture,
		}
	}
	return nil
}


func (r *inMemoryRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	fare, exist := r.rideFares[id]
	if !exist {
		return nil, fmt.Errorf("fare does not exist with ID: %s", id)
	}

	return fare, nil
}

func (r *inMemoryRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip

	return trip, nil
}

func (r *inMemoryRepository) SaveRideFare(ctx context.Context, f *domain.RideFareModel) error {
	r.rideFares[f.ID.Hex()] = f

	return nil
}

