package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
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

func (s *AppserverSubService) PgTypeToPb(aSub *qx.AppserverSub) *pb_appserversub.AppserverSub {
	return &pb_appserversub.AppserverSub{
		Id:          aSub.ID.String(),
		AppserverId: aSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(aSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aSub.UpdatedAt.Time),
	}
}

func (s *AppserverSubService) PgAppserverSubRowToPb(res *qx.GetUserAppserverSubsRow) *pb_appserversub.AppserverAndSub {
	appserver := &pb_appserver.Appserver{
		Id:        res.ID.String(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserversub.AppserverAndSub{
		Appserver: appserver,
		SubId:     res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) PgUserSubRowToPb(res *qx.GetAllUsersAppserverSubsRow) *pb_appserversub.AppuserAndSub {
	appuser := &pb_appuser.Appuser{
		Id:        res.ID.String(),
		Username:  res.Username,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserversub.AppuserAndSub{
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
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) ListAllUsersAppserverAndSub(
	appserverId uuid.UUID,
) ([]qx.GetAllUsersAppserverSubsRow, error) {

	aSubs, err := qx.New(s.dbConn).GetAllUsersAppserverSubs(s.ctx, appserverId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) DeleteByAppserver(id uuid.UUID) error {
	/* Removes a user from a server. */

	deleted, err := qx.New(s.dbConn).DeleteAppserverSub(s.ctx, id)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return nil
}
