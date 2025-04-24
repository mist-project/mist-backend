package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
)

type AppserverRoleService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

func NewAppserverRoleService(dbConn qx.DBTX, ctx context.Context) *AppserverRoleService {
	return &AppserverRoleService{dbConn: dbConn, ctx: ctx}
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
	appserverRole, err := qx.New(s.dbConn).CreateAppserverRole(s.ctx, obj)
	return &appserverRole, err
}

func (s *AppserverRoleService) ListAppserverRoles(appserverId uuid.UUID) ([]qx.AppserverRole, error) {
	aRoles, err := qx.New(s.dbConn).GetAppserverRoles(s.ctx, appserverId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aRoles, nil
}

func (s *AppserverRoleService) DeleteByAppserver(obj qx.DeleteAppserverRoleParams) error {
	deleted, err := qx.New(s.dbConn).DeleteAppserverRole(s.ctx, obj)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return nil
}
