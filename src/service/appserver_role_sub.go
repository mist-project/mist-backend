package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pb_server "mist/src/protos/server/v1"
	"mist/src/psql_db/qx"
)

type AppserverRoleSubService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverRoleSubService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverRoleSubService {
	return &AppserverRoleSubService{dbcPool: dbcPool, ctx: ctx}
}

func (service *AppserverRoleSubService) PgTypeToPb(appserverRole *qx.AppserverRoleSub) *pb_server.AppserverRoleSub {
	return &pb_server.AppserverRoleSub{
		Id:              appserverRole.ID.String(),
		AppserverRoleId: appserverRole.AppserverRoleID.String(),
		AppserverSubId:  appserverRole.AppserverSubID.String(),
	}
}

func (service *AppserverRoleSubService) Create(appserverRoleId string, appserverSubId string) (*qx.AppserverRoleSub, error) {
	validationErrors := []string{}
	if appserverRoleId == "" {
		validationErrors = AddValidationError("appserver_role_id", validationErrors)
	}

	if appserverSubId == "" {
		validationErrors = AddValidationError("appserver_sub_id", validationErrors)
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErrors, ", ")))
	}

	parsedAppserverRoleId, err := uuid.Parse(appserverRoleId)
	if err != nil {
		return nil, err
	}

	parsedAppserverSubId, err := uuid.Parse(appserverSubId)
	if err != nil {
		return nil, err
	}

	appserverRole, err := qx.New(service.dbcPool).CreateAppserverRoleSub(
		service.ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  parsedAppserverSubId,
			AppserverRoleID: parsedAppserverRoleId,
		},
	)
	return &appserverRole, err
}

func (service *AppserverRoleSubService) DeleteRoleSub(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	parsedOwnerUuid, err := uuid.Parse(ownerId)

	if err != nil {
		return err
	}

	deletedRows, err := qx.New(service.dbcPool).DeleteAppserverRoleSub(service.ctx, qx.DeleteAppserverRoleSubParams{
		ID: parsedUuid, OwnerID: parsedOwnerUuid,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deletedRows == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return nil
}
