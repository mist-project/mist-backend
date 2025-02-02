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

func (s *ChannelService) PgTypeToPb(c *qx.Channel) *pb_channel.Channel {
	return &pb_channel.Channel{
		Id:          c.ID.String(),
		Name:        c.Name,
		AppserverId: c.AppserverID.String(),
		CreatedAt:   timestamppb.New(c.CreatedAt.Time),
	}
}

func (s *ChannelService) Create(name string, appserverId string) (*qx.Channel, error) {
	validationErr := []string{}

	if name == "" {
		validationErr = AddValidationError("name", validationErr)
	}

	if appserverId == "" {
		validationErr = AddValidationError("appserver_id", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErr, ",")))
	}

	parsedUuid, err := uuid.Parse(appserverId)

	if err != nil {
		return nil, err
	}

	channel, err := qx.New(s.dbcPool).CreateChannel(
		s.ctx, qx.CreateChannelParams{Name: name, AppserverID: parsedUuid},
	)
	return &channel, err
}

func (s *ChannelService) GetById(id string) (*qx.Channel, error) {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	channel, err := qx.New(s.dbcPool).GetChannel(s.ctx, parsedUuid)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &channel, nil
}

func (s *ChannelService) List(name *wrappers.StringValue, appserverId *wrappers.StringValue) ([]qx.Channel, error) {
	// To query, remember to format the parameters
	var (
		fName pgtype.Text
		fAId  pgtype.UUID
	)

	if name != nil {
		fName = pgtype.Text{Valid: true, String: name.Value}
	}

	if appserverId != nil {
		parsedUuid, err := uuid.Parse(appserverId.Value)
		if err != nil {
			return nil, err
		}
		fAId = pgtype.UUID{Valid: true, Bytes: parsedUuid}
	} else {
		fAId = pgtype.UUID{Valid: false}
	}

	channels, err := qx.New(s.dbcPool).ListChannels(
		s.ctx, qx.ListChannelsParams{Name: fName, AppserverID: fAId},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return channels, nil
}

func (s *ChannelService) Delete(id string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	deleted, err := qx.New(s.dbcPool).DeleteChannel(s.ctx, parsedUuid)

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return err
}
