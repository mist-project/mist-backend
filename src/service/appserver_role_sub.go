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

func (s *AppserverRoleSubService) PgTypeToPb(arSub *qx.AppserverRoleSub) *pb_server.AppserverRoleSub {
	return &pb_server.AppserverRoleSub{
		Id:              arSub.ID.String(),
		AppserverRoleId: arSub.AppserverRoleID.String(),
		AppserverSubId:  arSub.AppserverSubID.String(),
	}
}

func (s *AppserverRoleSubService) Create(appserverRoleId string, appserverSubId string, appUserId string) (*qx.AppserverRoleSub, error) {
	validationErr := []string{}

	if appserverRoleId == "" {
		validationErr = AddValidationError("appserver_role_id", validationErr)
	}

	if appserverSubId == "" {
		validationErr = AddValidationError("appserver_sub_id", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): %s", ValidationError, strings.Join(validationErr, ", ")))
	}

	pARId, err := uuid.Parse(appserverRoleId)

	if err != nil {
		return nil, err
	}

	pASId, err := uuid.Parse(appserverSubId)
	if err != nil {
		return nil, err
	}

	pAUId, err := uuid.Parse(appUserId)
	if err != nil {
		return nil, err
	}

	appserverRole, err := qx.New(s.dbcPool).CreateAppserverRoleSub(
		s.ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  pASId,
			AppserverRoleID: pARId,
			AppUserID:       pAUId,
		},
	)

	return &appserverRole, err
}

func (s *AppserverRoleSubService) DeleteRoleSub(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	parsedOwnerUuid, err := uuid.Parse(ownerId)

	if err != nil {
		return err
	}

	deleted, err := qx.New(s.dbcPool).DeleteAppserverRoleSub(s.ctx, qx.DeleteAppserverRoleSubParams{
		ID: parsedUuid, AppUserID: parsedOwnerUuid,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return nil
}
