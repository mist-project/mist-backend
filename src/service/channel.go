package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type ChannelService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

// Creates a new ChannelService struct.
func NewChannelService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *ChannelService {
	return &ChannelService{ctx: ctx, dbConn: dbConn, db: db}
}

// Convert Channel db object to Channel protobuff object.
func (s *ChannelService) PgTypeToPb(c *qx.Channel) *pb_channel.Channel {
	return &pb_channel.Channel{
		Id:          c.ID.String(),
		Name:        c.Name,
		AppserverId: c.AppserverID.String(),
		CreatedAt:   timestamppb.New(c.CreatedAt.Time),
	}
}

// Creates a new appuser.
func (s *ChannelService) Create(obj qx.CreateChannelParams) (*qx.Channel, error) {
	channel, err := s.db.CreateChannel(s.ctx, obj)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) create channel error: %v", DatabaseError, err))
	}

	return &channel, err
}

// Gets an appserver detail by its id.
func (s *ChannelService) GetById(id uuid.UUID) (*qx.Channel, error) {
	channel, err := s.db.GetChannelById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
		}

		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return &channel, nil
}

// Lists all channels for an appserver. Name filter is also added but it may get deprecated.

func (s *ChannelService) List(obj qx.ListServerChannelsParams) ([]qx.Channel, error) {

	channels, err := s.db.ListServerChannels(s.ctx, obj)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return channels, nil
}

// Delete a channel object
func (s *ChannelService) Delete(id uuid.UUID) error {
	// TODO: add authorization layer before deleting
	deleted, err := s.db.DeleteChannel(s.ctx, id)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
	}

	return err
}
