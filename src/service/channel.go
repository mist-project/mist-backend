package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
)

type ChannelService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewChannelService(dbcPool *pgxpool.Pool, ctx context.Context) *ChannelService {
	return &ChannelService{dbcPool: dbcPool, ctx: ctx}
}

func (service *ChannelService) PgTypeToPb(channel *qx.Channel) *pb_channel.Channel {
	return &pb_channel.Channel{
		Id:          channel.ID.String(),
		Name:        channel.Name,
		AppserverId: channel.AppserverID.String(),
		CreatedAt:   timestamppb.New(channel.CreatedAt.Time),
	}
}

func (service *ChannelService) Create(name string, appserverId string) (*qx.Channel, error) {
	validationErrors := []string{}
	if name == "" {
		validationErrors = AddValidationError("name", validationErrors)
	}

	if appserverId == "" {
		validationErrors = AddValidationError("appserver_id", validationErrors)
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErrors, ",")))
	}

	parsedUuid, err := uuid.Parse(appserverId)

	if err != nil {
		return nil, err
	}

	channel, err := qx.New(service.dbcPool).CreateChannel(
		service.ctx, qx.CreateChannelParams{Name: name, AppserverID: parsedUuid},
	)
	return &channel, err
}

func (service *ChannelService) GetById(id string) (*qx.Channel, error) {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	channel, err := qx.New(service.dbcPool).GetChannel(service.ctx, parsedUuid)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &channel, nil
}

func (service *ChannelService) List(name *wrappers.StringValue, appserverId *wrappers.StringValue) ([]qx.Channel, error) {
	// To query, remember to format the parameters
	var formatName pgtype.Text
	var formatAppserverId pgtype.UUID
	if name != nil {
		formatName = pgtype.Text{Valid: true, String: name.Value}
	}

	if appserverId != nil {
		parsedUuid, err := uuid.Parse(appserverId.Value)
		if err != nil {
			return nil, err
		}
		formatAppserverId = pgtype.UUID{Valid: true, Bytes: parsedUuid}
	} else {
		formatAppserverId = pgtype.UUID{Valid: false}
	}

	channels, err := qx.New(service.dbcPool).ListChannels(
		service.ctx, qx.ListChannelsParams{Name: formatName, AppserverID: formatAppserverId},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return channels, nil
}

func (service *ChannelService) Delete(id string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	deletedRows, err := qx.New(service.dbcPool).DeleteChannel(service.ctx, parsedUuid)
	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deletedRows == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return err
}
