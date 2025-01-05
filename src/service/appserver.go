package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_server "mist/src/protos/server/v1"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverService {
	return &AppserverService{dbcPool: dbcPool, ctx: ctx}
}

func (service *AppserverService) PgTypeToPb(appserver *qx.Appserver) *pb_server.Appserver {
	return &pb_server.Appserver{
		Id:        appserver.ID.String(),
		Name:      appserver.Name,
		CreatedAt: timestamppb.New(appserver.CreatedAt.Time),
	}
}

func (service *AppserverService) Create(name string, userId string) (*qx.Appserver, error) {
	// Keeping the validationErrors variable as a way to show the pattern I'd like to follow (using a list of
	// validation errors to then send them)
	// Note: might change the pattern to use some sort of validation package. This might be duable by changing the
	// parameter in this method for example, to a struct type that can be validated. (Similar concept of python's
	// Pydantic object validation)
	validationErrors := []string{}
	if name == "" {
		validationErrors = AddValidationError("name", validationErrors)
	}

	if userId == "" {
		validationErrors = AddValidationError("user_id", validationErrors)
	}

	if len(validationErrors) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): missing name attribute", ValidationError))
	}

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): %v", ValidationError, err))
	}

	appserver, err := qx.New(service.dbcPool).CreateAppserver(service.ctx, qx.CreateAppserverParams{
		Name:    name,
		OwnerID: parsedUserId,
	})
	return &appserver, err
}

func (service *AppserverService) GetById(id string) (*qx.Appserver, error) {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	appserver, err := qx.New(service.dbcPool).GetAppserver(service.ctx, parsedUuid)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &appserver, nil
}

func (service *AppserverService) List(name *wrappers.StringValue, ownerId string) ([]qx.Appserver, error) {
	// To query remember do to: {"name": {"value": "boo"}}
	var formatName = pgtype.Text{Valid: false}
	if name != nil {
		formatName.Valid = true
		formatName.String = name.Value
	}

	parsedOwnerUuid, _ := uuid.Parse(ownerId)
	appservers, err := qx.New(service.dbcPool).ListUserAppservers(
		service.ctx, qx.ListUserAppserversParams{Name: formatName, OwnerID: parsedOwnerUuid},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appservers, nil
}

func (service *AppserverService) Delete(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	parsedOwnerUuid, _ := uuid.Parse(ownerId)

	deletedRows, err := qx.New(service.dbcPool).DeleteAppserver(
		service.ctx, qx.DeleteAppserverParams{ID: parsedUuid, OwnerID: parsedOwnerUuid})
	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deletedRows == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}
	return err
}
