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
	"mist/src/psql_db/qx"
)

type AppserverRoleService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverRoleService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverRoleService {
	return &AppserverRoleService{dbcPool: dbcPool, ctx: ctx}
}

func (s *AppserverRoleService) PgTypeToPb(aRole *qx.AppserverRole) *pb_appserver.AppserverRole {
	return &pb_appserver.AppserverRole{
		Id:          aRole.ID.String(),
		AppserverId: aRole.AppserverID.String(),
		Name:        aRole.Name,
		CreatedAt:   timestamppb.New(aRole.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aRole.UpdatedAt.Time),
	}
}

func (s *AppserverRoleService) Create(appserverId string, name string) (*qx.AppserverRole, error) {
	validationErr := []string{}

	if appserverId == "" {
		validationErr = AddValidationError("appserver_id", validationErr)
	}

	if name == "" {
		validationErr = AddValidationError("name", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErr, ", ")))
	}

	parsedAId, err := uuid.Parse(appserverId)

	if err != nil {
		return nil, err
	}

	appserverRole, err := qx.New(s.dbcPool).CreateAppserverRole(
		s.ctx, qx.CreateAppserverRoleParams{
			AppserverID: parsedAId,
			Name:        name,
		},
	)

	return &appserverRole, err
}

func (s *AppserverRoleService) ListAppserverRoles(appserverId string) ([]qx.AppserverRole, error) {
	parsedUuid, err := uuid.Parse(appserverId)
	if err != nil {
		return nil, err
	}

	aRoles, err := qx.New(s.dbcPool).GetAppserverRoles(
		s.ctx, parsedUuid,
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return aRoles, nil
}

func (s *AppserverRoleService) DeleteByAppserver(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	parsedOwnerUuid, err := uuid.Parse(ownerId)

	if err != nil {
		return err
	}

	deleted, err := qx.New(s.dbcPool).DeleteAppserverRole(s.ctx, qx.DeleteAppserverRoleParams{
		ID: parsedUuid, AppuserID: parsedOwnerUuid,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return nil
}
