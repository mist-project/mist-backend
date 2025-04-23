package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

func NewAppserverSubService(dbConn qx.DBTX, ctx context.Context) *AppserverSubService {
	return &AppserverSubService{dbConn: dbConn, ctx: ctx}
}

func (s *AppserverSubService) PgTypeToPb(aSub *qx.AppserverSub) *pb_appserver.AppserverSub {
	return &pb_appserver.AppserverSub{
		Id:          aSub.ID.String(),
		AppserverId: aSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(aSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aSub.UpdatedAt.Time),
	}
}

func (s *AppserverSubService) PgAppserverSubRowToPb(res *qx.GetUserAppserverSubsRow) *pb_appserver.AppserverAndSub {
	appserver := &pb_appserver.Appserver{
		Id:        res.ID.String(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserver.AppserverAndSub{
		Appserver: appserver,
		SubId:     res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) PgUserSubRowToPb(res *qx.GetAllUsersAppserverSubsRow) *pb_appserver.AppuserAndSub {
	appuser := &pb_appuser.Appuser{
		Id:        res.ID.String(),
		Username:  res.Username,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserver.AppuserAndSub{
		Appuser: appuser,
		SubId:   res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) Create(obj qx.CreateAppserverSubParams) (*qx.AppserverSub, error) {
	appserverSub, err := qx.New(s.dbConn).CreateAppserverSub(s.ctx, obj)
	return &appserverSub, err
}

func (s *AppserverSubService) ListUserAppserverAndSub(userId uuid.UUID) ([]qx.GetUserAppserverSubsRow, error) {
	/* Returns all servers a user belongs to. */

	aSubs, err := qx.New(s.dbConn).GetUserAppserverSubs(s.ctx, userId)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) ListAllUsersAppserverAndSub(
	appserverId uuid.UUID,
) ([]qx.GetAllUsersAppserverSubsRow, error) {

	aSubs, err := qx.New(s.dbConn).GetAllUsersAppserverSubs(s.ctx, appserverId)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) DeleteByAppserver(id uuid.UUID) error {
	/* Removes a user from a server. */

	deleted, err := qx.New(s.dbConn).DeleteAppserverSub(s.ctx, id)

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return nil
}
