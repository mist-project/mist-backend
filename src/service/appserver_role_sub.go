package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
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

func (s *AppserverRoleSubService) PgTypeToPb(arSub *qx.AppserverRoleSub) *pb_appserverrolesub.AppserverRoleSub {
	return &pb_appserverrolesub.AppserverRoleSub{
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
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return &appserverRole, err
}

// Get all the roles each user has in a server.
func (s *AppserverRoleSubService) ListServerRoleSubs(
	appserverId uuid.UUID,
) ([]qx.ListServerRoleSubsRow, error) {

	rows, err := s.db.ListServerRoleSubs(s.ctx, appserverId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return rows, nil
}

// Removes a role to a particular user.
func (s *AppserverRoleSubService) Delete(obj qx.DeleteAppserverRoleSubParams) error {
	deleted, err := s.db.DeleteAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
	}

	return nil
}
