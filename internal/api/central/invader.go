package central

import (
	"context"

	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) RetrieveInvaderPicture(
	ctx context.Context, request *centralpb.RetrieveInvaderPictureRequest,
) (*centralpb.RetrieveInvaderPictureReply, error) {
	if request.GetEventId() == "" || request.GetInvaderId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing event or invader ID")
	}

	eventID, err := primitive.ObjectIDFromHex(request.GetEventId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event ID: %v", err)
	}
	invaderID, err := primitive.ObjectIDFromHex(request.GetInvaderId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid invader ID: %v", err)
	}

	picture, err := s.db.GetInvaderPicture(ctx, &database.GetInvaderPictureRequest{
		EventID:   eventID,
		InvaderID: invaderID,
	})
}
