package main

import (
	"context"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/domain"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/repository"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/service"
)

func main() {
	ctx := context.Background()


	inMemoryRepo := repository.NewInMemoryRepository()

	serv := service.NewService(inMemoryRepo)

	fare := &domain.RideFareModel{
		UserID: "23",
	}

	serv.CreateTrip(ctx, fare)
}