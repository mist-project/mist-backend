package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/qx"
)

type ChannelRoleService struct {
	ctx  context.Context
	deps *ServiceDeps
}

func NewChannelRoleService(ctx context.Context, deps *ServiceDeps) *ChannelRoleService {
	return &ChannelRoleService{ctx: ctx, deps: deps}
}

func (s *ChannelRoleService) PgTypeToPb(cRole *qx.ChannelRole) *channel_role.ChannelRole {
	return &channel_role.ChannelRole{
		Id:              cRole.ID.String(),
		ChannelId:       cRole.ChannelID.String(),
		AppserverRoleId: cRole.AppserverRoleID.String(),
		CreatedAt:       timestamppb.New(cRole.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(cRole.UpdatedAt.Time),
	}
}

// Creates an appserver role.
func (s *ChannelRoleService) Create(obj qx.CreateChannelRoleParams) (*qx.ChannelRole, error) {
	channelRole, err := s.deps.Db.CreateChannelRole(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	NewChannelService(s.ctx, s.deps).SendChannelListingUpdateNotificationToUsers(
		nil, channelRole.AppserverID,
	)

	return &channelRole, err
}

// Lists all the roles for an appserver.
func (s *ChannelRoleService) ListChannelRoles(channelId uuid.UUID) ([]qx.ChannelRole, error) {
	cRoles, err := s.deps.Db.ListChannelRoles(s.ctx, channelId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return cRoles, nil
}

// Gets an appserver role by its id.
func (s *ChannelRoleService) GetById(id uuid.UUID) (*qx.ChannelRole, error) {
	role, err := s.deps.Db.GetChannelRoleById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError(fmt.Sprintf("unable to find channel role with id: %v", id), slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &role, nil
}

// Deletes a role from a server, only owner of server and delete role
func (s *ChannelRoleService) Delete(id uuid.UUID) error {

	channelRole, err := s.GetById(id) // Check if the role exists

	if err != nil {
		return faults.ExtendError(err)
	}

	deleted, err := s.deps.Db.DeleteChannelRole(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("unable to find channel role with id: %v", id), slog.LevelDebug)
	}

	NewChannelService(s.ctx, s.deps).SendChannelListingUpdateNotificationToUsers(
		nil, channelRole.AppserverID,
	)

	return nil
}
