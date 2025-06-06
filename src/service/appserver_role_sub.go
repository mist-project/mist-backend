package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/producer"
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverRoleSubService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
	mp     producer.MessageProducer
}

func NewAppserverRoleSubService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier, mp producer.MessageProducer) *AppserverRoleSubService {
	return &AppserverRoleSubService{ctx: ctx, dbConn: dbConn, db: db, mp: mp}
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
	appserverRole, err := s.db.CreateAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	NewChannelService(s.ctx, s.dbConn, s.db, s.mp).SendChannelListingUpdateNotificationToUsers(
		&qx.Appuser{ID: appserverRole.AppuserID},
		appserverRole.AppserverID,
	)

	return &appserverRole, err
}

// Get all the roles each user has in a server.
func (s *AppserverRoleSubService) ListServerRoleSubs(
	appserverId uuid.UUID,
) ([]qx.ListServerRoleSubsRow, error) {

	rows, err := s.db.ListServerRoleSubs(s.ctx, appserverId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return rows, nil
}

// Gets an appserver role sub by its id.
func (s *AppserverRoleSubService) GetById(id uuid.UUID) (*qx.AppserverRoleSub, error) {
	role, err := s.db.GetAppserverRoleSubById(s.ctx, id)

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

	deleted, err := s.db.DeleteAppserverRoleSub(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("no appserver role sub found for id: %s", id), slog.LevelDebug)
	}

	NewChannelService(s.ctx, s.dbConn, s.db, s.mp).SendChannelListingUpdateNotificationToUsers(
		&qx.Appuser{ID: roleSub.AppuserID},
		roleSub.AppserverID,
	)

	return nil
}
