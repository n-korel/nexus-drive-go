package domain

import (
	pb "github.com/n-korel/nexus-drive-go/shared/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID primitive.ObjectID
	UserID string
	PackageSlug string // van, luxury, sedan
	TotalPriceInCents float64
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return  &pb.RideFare{
		Id: r.ID.Hex(),
		UserID: r.UserID,
		PackageSlug: r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	var protoFares []*pb.RideFare
	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}

	return protoFares
}