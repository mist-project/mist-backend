package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
)

type AppserverRoleSubService struct {
	ctx  context.Context
	deps *ServiceDeps
}

func NewAppserverRoleSubService(ctx context.Context, deps *ServiceDeps) *AppserverRoleSubService {
	return &AppserverRoleSubService{ctx: ctx, deps: deps}
}

func (s *AppserverRoleSubService) PgTypeToPb(arSub *qx.AppserverRoleSub) *appserver_role_sub.AppserverRoleSub {
	return &appserver_role_sub.AppserverRoleSub{
		Id:              arSub.ID.String(),
		AppserverRoleId: arSub.AppserverRoleID.String(),
		AppuserId:       arSub.AppuserID.String(),
		AppserverId:     arSub.AppserverID.String(),
	}
}

// Adds a server role to a user.
func (s *AppserverRoleSubService) Create(obj qx.CreateAppserverRoleSubParams) (*qx.AppserverRoleSub, error) {
	appserverRole, err := s.deps.Db.CreateAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	NewChannelService(s.ctx, s.deps).SendChannelListingUpdateNotificationToUsers(
		&qx.Appuser{ID: appserverRole.AppuserID},
		appserverRole.AppserverID,
	)

	return &appserverRole, err
}

// Get all the roles each user has in a server.
func (s *AppserverRoleSubService) ListServerRoleSubs(
	appserverId uuid.UUID,
) ([]qx.ListServerRoleSubsRow, error) {

	rows, err := s.deps.Db.ListServerRoleSubs(s.ctx, appserverId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return rows, nil
}

// Gets an appserver role sub by its id.
func (s *AppserverRoleSubService) GetById(id uuid.UUID) (*qx.AppserverRoleSub, error) {
	role, err := s.deps.Db.GetAppserverRoleSubById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError(fmt.Sprintf("no appserver role sub found for id: %s", id), slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &role, nil
}

// Removes a role to a particular user.
func (s *AppserverRoleSubService) Delete(id uuid.UUID) error {
	roleSub, err := s.GetById(id)

	if err != nil {
		return faults.ExtendError(err)
	}

	deleted, err := s.deps.Db.DeleteAppserverRoleSub(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("no appserver role sub found for id: %s", id), slog.LevelDebug)
	}

	NewChannelService(s.ctx, s.deps).SendChannelListingUpdateNotificationToUsers(
		&qx.Appuser{ID: roleSub.AppuserID},
		roleSub.AppserverID,
	)

	return nil
}
