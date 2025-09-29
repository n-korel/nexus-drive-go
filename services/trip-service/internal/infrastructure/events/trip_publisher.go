package events

import (
	"context"
	"encoding/json"

	"github.com/n-korel/nexus-drive-go/services/trip-service/internal/domain"
	"github.com/n-korel/nexus-drive-go/shared/contracts"
	"github.com/n-korel/nexus-drive-go/shared/messaging"
)


type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	payload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	tripEventJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data: tripEventJSON,
	})
}