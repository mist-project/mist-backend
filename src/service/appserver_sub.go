package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverSubService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverSubService {
	return &AppserverSubService{dbcPool: dbcPool, ctx: ctx}
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

func (s *AppserverSubService) Create(appserverId string, ownerId string) (*qx.AppserverSub, error) {
	validationErr := []string{}

	if appserverId == "" {
		validationErr = AddValidationError("appserver_id", validationErr)
	}

	if ownerId == "" {
		validationErr = AddValidationError("app_user_id", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErr, ", ")))
	}

	pAId, err := uuid.Parse(appserverId)

	if err != nil {
		return nil, err
	}

	pUId, err := uuid.Parse(ownerId)
	if err != nil {
		return nil, err
	}

	appserverSub, err := qx.New(s.dbcPool).CreateAppserverSub(
		s.ctx, qx.CreateAppserverSubParams{
			AppserverID: pAId,
			AppuserID:   pUId,
		},
	)

	return &appserverSub, err
}

func (s *AppserverSubService) ListUserAppserverAndSub(userId string) ([]qx.GetUserAppserverSubsRow, error) {
	/* Returns all servers a user belongs to. */

	parsedUuid, err := uuid.Parse(userId)

	if err != nil {
		return nil, err
	}

	aSubs, err := qx.New(s.dbcPool).GetUserAppserverSubs(
		s.ctx, parsedUuid,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) ListAllUsersAppserverAndSub(
	appserverId string,
) ([]qx.GetAllUsersAppserverSubsRow, error) {
	/* Returns all users in a server. */
	parsedUuid, err := uuid.Parse(appserverId)

	if err != nil {
		return nil, err
	}

	aSubs, err := qx.New(s.dbcPool).GetAllUsersAppserverSubs(
		s.ctx, parsedUuid,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aSubs, nil
}

func (s *AppserverSubService) DeleteByAppserver(id string) error {
	/* Removes a user from a server. */
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	deleted, err := qx.New(s.dbcPool).DeleteAppserverSub(s.ctx, parsedUuid)
	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return nil
}
