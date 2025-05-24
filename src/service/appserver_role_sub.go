package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mist/src/errors/message"
	pb_appserver_role_sub "mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverRoleSubService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverRoleSubService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverRoleSubService {
	return &AppserverRoleSubService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverRoleSubService) PgTypeToPb(arSub *qx.AppserverRoleSub) *pb_appserver_role_sub.AppserverRoleSub {
	return &pb_appserver_role_sub.AppserverRoleSub{
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
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &appserverRole, err
}

// Get all the roles each user has in a server.
func (s *AppserverRoleSubService) ListServerRoleSubs(
	appserverId uuid.UUID,
) ([]qx.ListServerRoleSubsRow, error) {

	rows, err := s.db.ListServerRoleSubs(s.ctx, appserverId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return rows, nil
}

// Gets an appserver role sub by its id.
func (s *AppserverRoleSubService) GetById(id uuid.UUID) (*qx.AppserverRoleSub, error) {
	role, err := s.db.GetAppserverRoleSubById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, message.NotFoundError(message.NotFound)
		}

		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &role, nil
}

// Removes a role to a particular user.
func (s *AppserverRoleSubService) Delete(obj qx.DeleteAppserverRoleSubParams) error {
	deleted, err := s.db.DeleteAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError("resource not found")
	}

	return nil
}
