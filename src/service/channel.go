package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type ChannelService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
	p      producer.MessageProducer
}

// Creates a new ChannelService struct.
func NewChannelService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier, p producer.MessageProducer) *ChannelService {
	return &ChannelService{ctx: ctx, dbConn: dbConn, db: db, p: p}
}

// Convert Channel db object to Channel protobuff object.
func (s *ChannelService) PgTypeToPb(c *qx.Channel) *channel.Channel {
	return &channel.Channel{
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
		return nil, message.DatabaseError(fmt.Sprintf("create channel error: %v", err))
	}

	err = s.p.SendMessage(s.PgTypeToPb(&channel), event.ActionType_ACTION_ADD_CHANNEL, nil)

	if err != nil {
		// TODO: send error to some other place to handle it
		fmt.Println(err)
		err = nil
	}

	return &channel, err
}

// Gets an appserver detail by its id.
func (s *ChannelService) GetById(id uuid.UUID) (*qx.Channel, error) {
	channel, err := s.db.GetChannelById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, message.NotFoundError(message.NotFound)
		}

		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &channel, nil
}

// Lists all channels for an appserver. Name filter is also added but it may get deprecated.
func (s *ChannelService) ListServerChannels(obj qx.ListServerChannelsParams) ([]qx.Channel, error) {

	channels, err := s.db.ListServerChannels(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return channels, nil
}

// Delete a channel object
func (s *ChannelService) Delete(id uuid.UUID) error {
	// TODO: add authorization layer before deleting
	deleted, err := s.db.DeleteChannel(s.ctx, id)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}

	return err
}
