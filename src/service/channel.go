package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
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
		IsPrivate:   c.IsPrivate,
		AppserverId: c.AppserverID.String(),
		CreatedAt:   timestamppb.New(c.CreatedAt.Time),
	}
}

// Creates a new appuser.
func (s *ChannelService) Create(obj qx.CreateChannelParams) (*qx.Channel, error) {

	channel, err := s.db.CreateChannel(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("create channel error: %v", err), slog.LevelError)
	}

	// Send notification to all users in the channel
	s.SendChannelListingUpdateNotificationToUsers(nil, channel.AppserverID)

	return &channel, err
}

// Gets an appserver detail by its id.
func (s *ChannelService) GetById(id uuid.UUID) (*qx.Channel, error) {
	channel, err := s.db.GetChannelById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError("channel not found", slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &channel, nil
}

// Lists all channels for an appserver. Name filter is also added but it may get deprecated.
func (s *ChannelService) ListServerChannels(obj qx.ListServerChannelsParams) ([]qx.Channel, error) {

	// TODO: This should only return channel that the user has access to. Pull the channels which user has roles to
	// and pulls all the channels without roles in the server.
	channels, err := s.db.ListServerChannels(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return channels, nil
}

// Lists all channels for an appserver. Name filter is also added but it may get deprecated.
func (s *ChannelService) Filter(obj qx.FilterChannelParams) ([]qx.Channel, error) {

	// TODO: This should only return channel that the user has access to. Pull the channels which user has roles to
	// and pulls all the channels without roles in the server.
	channels, err := s.db.FilterChannel(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return channels, nil
}

// Delete a channel object
func (s *ChannelService) Delete(id uuid.UUID) error {
	// TODO: doing double queries here "fetching" the sub and then deleting it. maybe change this so that
	// we can do it in one query.
	channel, err := s.db.GetChannelById(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("error deleting channel: %v", err), slog.LevelError)
	}

	deleted, err := s.db.DeleteChannel(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("error deleting channel: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("unable to find channel with id: (%v)", id), slog.LevelDebug)
	}

	s.SendChannelListingUpdateNotificationToUsers(nil, channel.AppserverID)

	return err
}

func (s *ChannelService) SendChannelListingUpdateNotificationToUsers(u *qx.Appuser, appserverId uuid.UUID) {
	var (
		appuserIds []uuid.UUID
	)

	if u != nil {
		appuserIds = []uuid.UUID{u.ID}
	} else {
		// get all users in the appserver
		appusers, err := s.db.ListAppserverUserSubs(s.ctx, appserverId)

		if err != nil {
			faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError).LogError(s.ctx)
			return
		}

		// if no users, early exit
		if len(appusers) == 0 {
			return
		}

		appuserIds = make([]uuid.UUID, 0, len(appusers))

		// collect all appuser ids
		for _, user := range appusers {
			appuserIds = append(appuserIds, user.AppuserID)
		}
	}

	// get all available channel to each user
	channelUsers, err := s.db.GetChannelsForUsers(
		s.ctx, qx.GetChannelsForUsersParams{Column1: appuserIds, AppserverID: appserverId},
	)

	if err != nil {
		faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError).LogError(s.ctx)
		return
	}

	userChannelMap := make(map[uuid.UUID][]*channel.Channel)

	// map user ids to their channels
	for _, cu := range channelUsers {
		userChannelMap[cu.AppuserID] = append(userChannelMap[cu.AppuserID], &channel.Channel{
			Id:          cu.ChannelID.String(),
			Name:        cu.ChannelName.String,
			AppserverId: cu.ChannelAppserverID.String(),
			IsPrivate:   cu.ChannelIsPrivate.Bool,
		})
	}

	// if no channels, early exit
	if len(userChannelMap) == 0 {
		return
	}

	for userId, channels := range userChannelMap {
		s.mp.SendMessage(channels, event.ActionType_ACTION_LIST_CHANNELS, []*appuser.Appuser{{Id: userId.String()}})
	}
}
