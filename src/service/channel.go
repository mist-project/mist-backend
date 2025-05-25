package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	"mist/src/producer"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type ChannelService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
	mp     producer.MessageProducer
}

// Creates a new ChannelService struct.
func NewChannelService(
	ctx context.Context, dbConn *pgxpool.Pool, db db.Querier, mp producer.MessageProducer,
) *ChannelService {
	return &ChannelService{ctx: ctx, dbConn: dbConn, db: db, mp: mp}
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

	// Send notification to all users in the channel
	s.sendNotificationToChannelUsers(&channel, s.PgTypeToPb(&channel), event.ActionType_ACTION_ADD_CHANNEL)

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
	// TODO: doing double queries here "fetching" the sub and then deleting it. maybe change this so that
	// we can do it in one query.
	channel, subErr := s.db.GetChannelById(s.ctx, id)
	deleted, err := s.db.DeleteChannel(s.ctx, id)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}

	if subErr == nil {
		s.sendNotificationToChannelUsers(&channel, s.PgTypeToPb(&channel), event.ActionType_ACTION_REMOVE_CHANNEL)
	}

	return err
}

func (s *ChannelService) sendNotificationToChannelUsers(channel *qx.Channel, pbC *channel.Channel, action event.ActionType) {
	var (
		err   error
		users []*appuser.Appuser
	)

	roles, err := s.db.FilterChannelRole(s.ctx, qx.FilterChannelRoleParams{
		ChannelID: pgtype.UUID{Bytes: channel.ID, Valid: true},
	})

	if err != nil {
		s.mp.NotifyMessageFailure(fmt.Errorf("(ChannelService|%v) error getting channel roles: %v", action, err))
		return
	}

	// If there are roles in the channel, only users with those roles will be notified
	if len(roles) > 0 {
		// Extract user IDs from roles
		userIDs := make([]uuid.UUID, 0, len(roles))
		for _, role := range roles {
			userIDs = append(userIDs, role.ID)
		}

		// Get appusers by roles in the channel
		appusers, err := s.db.GetChannelUsersByRoles(s.ctx, userIDs)
		if err != nil {
			s.mp.NotifyMessageFailure(fmt.Errorf("(ChannelService|%v) error getting channel user by roles: %v", action, err))
			return
		}
		users = make([]*appuser.Appuser, 0, len(appusers))

		for _, u := range appusers {
			users = append(users, &appuser.Appuser{
				Id:       u.ID.String(),
				Username: u.Username,
			})
		}

	} else {
		// No roles in the channel, so all users have access to the channel
		userSubs, err := NewAppserverSubService(s.ctx, s.dbConn, s.db, s.mp).ListAppserverUserSubs(channel.AppserverID)

		if err != nil {
			s.mp.NotifyMessageFailure(fmt.Errorf("(ChannelService|%v) error getting appserver subs: %v", action, err))
			return
		}
		users = make([]*appuser.Appuser, 0, len(userSubs))

		for _, sub := range userSubs {
			users = append(users, &appuser.Appuser{Id: sub.ID.String(), Username: sub.Username})
		}
	}

	s.mp.SendMessage(pbC, action, users)
}
