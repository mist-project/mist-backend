package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
)

type ChannelService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

// Creates a new ChannelService struct
func NewChannelService(dbConn qx.DBTX, ctx context.Context) *ChannelService {
	return &ChannelService{dbConn: dbConn, ctx: ctx}
}

func (s *ChannelService) PgTypeToPb(c *qx.Channel) *pb_channel.Channel {
	return &pb_channel.Channel{
		Id:          c.ID.String(),
		Name:        c.Name,
		AppserverId: c.AppserverID.String(),
		CreatedAt:   timestamppb.New(c.CreatedAt.Time),
	}
}

func (s *ChannelService) Create(obj qx.CreateChannelParams) (*qx.Channel, error) {
	channel, err := qx.New(s.dbConn).CreateChannel(s.ctx, obj)
	return &channel, err
}

func (s *ChannelService) GetById(id uuid.UUID) (*qx.Channel, error) {
	channel, err := qx.New(s.dbConn).GetChannelById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
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

	channels, err := qx.New(s.dbConn).ListChannels(
		s.ctx, qx.ListChannelsParams{Name: fName, AppserverID: fAId},
	)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return channels, nil
}

func (s *ChannelService) Delete(id uuid.UUID) error {
	// TODO: add authorization layer before deleting
	deleted, err := qx.New(s.dbConn).DeleteChannel(s.ctx, id)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return err
}
