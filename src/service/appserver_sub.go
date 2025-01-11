package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_server "mist/src/protos/server/v1"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverSubService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverSubService {
	return &AppserverSubService{dbcPool: dbcPool, ctx: ctx}
}

func (service *AppserverSubService) PgTypeToPb(appserverSub *qx.AppserverSub) *pb_server.AppserverSub {
	return &pb_server.AppserverSub{
		Id:          appserverSub.ID.String(),
		AppserverId: appserverSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(appserverSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(appserverSub.UpdatedAt.Time),
	}
}

func (service *AppserverSubService) PgUserSubRowToPb(results *qx.GetUserAppserverSubsRow) *pb_server.AppserverAndSub {
	appserver := &pb_server.Appserver{
		Id:        results.ID.String(),
		Name:      results.Name,
		CreatedAt: timestamppb.New(results.CreatedAt.Time),
		UpdatedAt: timestamppb.New(results.UpdatedAt.Time),
	}
	return &pb_server.AppserverAndSub{
		Appserver: appserver,
		SubId:     results.AppserverSubID.String(),
	}
}

func (service *AppserverSubService) Create(appserverId string, ownerId string) (*qx.AppserverSub, error) {
	validationErrors := []string{}
	if appserverId == "" {
		validationErrors = AddValidationError("appserver_id", validationErrors)
	}

	if ownerId == "" {
		validationErrors = AddValidationError("app_user_id", validationErrors)
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErrors, ", ")))
	}

	parsedAppserverId, err := uuid.Parse(appserverId)
	if err != nil {
		return nil, err
	}

	parsedUserId, err := uuid.Parse(ownerId)
	if err != nil {
		return nil, err
	}

	appserverSub, err := qx.New(service.dbcPool).CreateAppserverSub(
		service.ctx, qx.CreateAppserverSubParams{
			AppserverID: parsedAppserverId,
			AppUserID:   parsedUserId,
		},
	)
	return &appserverSub, err
}

func (service *AppserverSubService) ListUserAppserverAndSub(ownerId string) ([]qx.GetUserAppserverSubsRow, error) {
	// TODO TOMORROW: REPLACE THIS QUERY WE DONT NEED TO FILTER BY APPSERVER
	parsedUuid, err := uuid.Parse(ownerId)
	if err != nil {
		return nil, err
	}

	appserverSubs, err := qx.New(service.dbcPool).GetUserAppserverSubs(
		service.ctx, parsedUuid,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appserverSubs, nil
}

func (service *AppserverSubService) DeleteByAppserver(id string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	deletedRows, err := qx.New(service.dbcPool).DeleteAppserverSub(service.ctx, parsedUuid)
	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deletedRows == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return nil
}
