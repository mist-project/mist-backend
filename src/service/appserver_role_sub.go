package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
)

type AppserverRoleSubService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

func NewAppserverRoleSubService(dbConn qx.DBTX, ctx context.Context) *AppserverRoleSubService {
	return &AppserverRoleSubService{dbConn: dbConn, ctx: ctx}
}

func (s *AppserverRoleSubService) PgTypeToPb(arSub *qx.AppserverRoleSub) *pb_appserverrolesub.AppserverRoleSub {
	return &pb_appserverrolesub.AppserverRoleSub{
		Id:              arSub.ID.String(),
		AppserverRoleId: arSub.AppserverRoleID.String(),
		AppuserId:       arSub.AppuserID.String(),
		AppserverId:     arSub.AppserverID.String(),
	}
}

func (s *AppserverRoleSubService) Create(
	obj qx.CreateAppserverRoleSubParams,
) (*qx.AppserverRoleSub, error) {
	appserverRole, err := qx.New(s.dbConn).CreateAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &appserverRole, err
}

func (s *AppserverRoleSubService) GetAppserverAllUserRoleSubs(
	appserverId uuid.UUID,
) ([]qx.GetAppserverAllUserRoleSubsRow, error) {

	rows, err := qx.New(s.dbConn).GetAppserverAllUserRoleSubs(s.ctx, appserverId)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return rows, nil
}

func (s *AppserverRoleSubService) DeleteRoleSub(obj qx.DeleteAppserverRoleSubParams) error {
	deleted, err := qx.New(s.dbConn).DeleteAppserverRoleSub(s.ctx, obj)

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return nil
}
