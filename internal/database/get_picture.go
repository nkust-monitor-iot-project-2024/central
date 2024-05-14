package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/slogext"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

var (
	ErrPictureNotFound   = errors.New("picture not found")
	ErrFileSizeUnmatched = errors.New("downloaded file size mismatch")
)

func (d *database) GetMovementPicture(ctx context.Context, req *GetMovementPictureRequest) (*GetMovementPictureResponse, error) {
	return d.getPicture(ctx, getMovementPath(req.EventID))
}

type GetMovementPictureRequest struct {
	EventID primitive.ObjectID
}

type GetMovementPictureResponse = GetPictureResponse

func (d *database) GetInvaderPicture(ctx context.Context, req *GetInvaderPictureRequest) (*GetInvaderPictureResponse, error) {
	return d.getPicture(ctx, getInvaderPath(req.EventID, req.InvaderID))
}

type GetInvaderPictureRequest struct {
	EventID   primitive.ObjectID
	InvaderID primitive.ObjectID
}

type GetInvaderPictureResponse = GetPictureResponse

func (d *database) getPicture(ctx context.Context, path string) (*GetPictureResponse, error) {
	buffer := bytes.Buffer{}

	size, err := d.Fs().DownloadToStreamByName(path, &buffer)
	if err != nil {
		if errors.Is(err, gridfs.ErrFileNotFound) {
			return nil, ErrPictureNotFound
		}

		d.logger.ErrorContext(ctx, "failed to download invader picture", slogext.Error(err))
		return nil, fmt.Errorf("download picture: %w", err)
	}

	if size != int64(buffer.Len()) {
		d.logger.ErrorContext(
			ctx, "downloaded file size mismatch",
			slog.Int64("expected", size), slog.Int64("actual", int64(buffer.Len())),
			slog.String("path", path),
		)
		return nil, ErrFileSizeUnmatched
	}

	return &GetPictureResponse{
		Picture: buffer.Bytes(),
	}, nil
}

type GetPictureResponse struct {
	Picture []byte
}
