package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_servers "mist/src/protos/server/v1"
	"mist/src/psql_db/qx"
)

type AppserverRoleService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverRoleService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverRoleService {
	return &AppserverRoleService{dbcPool: dbcPool, ctx: ctx}
}

func (service *AppserverRoleService) PgTypeToPb(appserverRole *qx.AppserverRole) *pb_servers.AppserverRole {
	return &pb_servers.AppserverRole{
		Id:          appserverRole.ID.String(),
		AppserverId: appserverRole.AppserverID.String(),
		Name:        appserverRole.Name,
		CreatedAt:   timestamppb.New(appserverRole.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(appserverRole.UpdatedAt.Time),
	}
}

func (service *AppserverRoleService) Create(appserverId string, name string) (*qx.AppserverRole, error) {
	validationErrors := []string{}
	if appserverId == "" {
		validationErrors = AddValidationError("appserver_id", validationErrors)
	}

	if name == "" {
		validationErrors = AddValidationError("name", validationErrors)
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErrors, ", ")))
	}

	parsedAppserverId, err := uuid.Parse(appserverId)
	if err != nil {
		return nil, err
	}

	appserverRole, err := qx.New(service.dbcPool).CreateAppserverRole(
		service.ctx, qx.CreateAppserverRoleParams{
			AppserverID: parsedAppserverId,
			Name:        name,
		},
	)
	return &appserverRole, err
}

func (service *AppserverRoleService) ListAppserverRoles(ownerId string) ([]qx.AppserverRole, error) {
	// TODO TOMORROW: REPLACE THIS QUERY WE DONT NEED TO FILTER BY APPSERVER
	parsedUuid, err := uuid.Parse(ownerId)
	if err != nil {
		return nil, err
	}

	appserverRoles, err := qx.New(service.dbcPool).GetAppserverRoles(
		service.ctx, parsedUuid,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appserverRoles, nil
}

func (service *AppserverRoleService) DeleteByAppserver(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	parsedOwnerUuid, err := uuid.Parse(ownerId)

	if err != nil {
		return err
	}

	deletedRows, err := qx.New(service.dbcPool).DeleteAppserverRole(service.ctx, qx.DeleteAppserverRoleParams{
		ID: parsedUuid, OwnerID: parsedOwnerUuid,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deletedRows == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return nil
}
