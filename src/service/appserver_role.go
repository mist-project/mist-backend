package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverRoleService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverRoleService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverRoleService {
	return &AppserverRoleService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverRoleService) PgTypeToPb(aRole *qx.AppserverRole) *pb_appserverrole.AppserverRole {
	return &pb_appserverrole.AppserverRole{
		Id:          aRole.ID.String(),
		AppserverId: aRole.AppserverID.String(),
		Name:        aRole.Name,
		CreatedAt:   timestamppb.New(aRole.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aRole.UpdatedAt.Time),
	}
}

func (s *AppserverRoleService) Create(obj qx.CreateAppserverRoleParams) (*qx.AppserverRole, error) {
	appserverRole, err := s.db.CreateAppserverRole(s.ctx, obj)
	return &appserverRole, err
}

func (s *AppserverRoleService) ListAppserverRoles(appserverId uuid.UUID) ([]qx.AppserverRole, error) {
	aRoles, err := s.db.GetAppserverRoles(s.ctx, appserverId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return aRoles, nil
}

func (s *AppserverRoleService) DeleteByAppserver(obj qx.DeleteAppserverRoleParams) error {
	deleted, err := s.db.DeleteAppserverRole(s.ctx, obj)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
	}
	return nil
}
