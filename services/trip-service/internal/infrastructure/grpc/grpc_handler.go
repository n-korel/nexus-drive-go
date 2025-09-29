package grpc

import (
	"context"
	"log"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/domain"
	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/infrastructure/events"
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
	"github.com/n-korel/nexus-drive-go/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer

	service domain.TripService
	publisher *events.TripEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	// Fetch and validate fare
	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate fare: %v", err)
	}

	// Create Trip
	trip, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}


	if err := h.publisher.PublishTripCreated(ctx, trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish trip created event: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}


func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()


	pickupCoordinate := &types.Coordinate{
		Latitude: pickup.Latitude,
		Longitude: pickup.Longitude,
	}

	destinationCoordinate := &types.Coordinate{
		Latitude: destination.Latitude,
		Longitude: destination.Longitude,
	}

	userID := req.GetUserID()

	route, err := h.service.GetRoute(ctx, pickupCoordinate, destinationCoordinate)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate ride fares: %v", err)
	}


	return &pb.PreviewTripResponse{
		Route: route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}