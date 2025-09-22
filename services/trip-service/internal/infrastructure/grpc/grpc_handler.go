package grpc

import (
	"context"
	"log"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/domain"
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
	"github.com/n-korel/nexus-drive-go/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer

	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
		return nil, status.Errorf(codes.Unimplemented, "method CreateTrip not implemented")
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

	t, err := h.service.GetRoute(ctx, pickupCoordinate, destinationCoordinate)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	estimatedFares := h.service.EstimatePackagesPriceWithRoute(t)

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate ride fares: %v", err)
	}


	return &pb.PreviewTripResponse{
		Route: t.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}