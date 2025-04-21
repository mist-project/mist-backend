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

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppserverService(dbcPool *pgxpool.Pool, ctx context.Context) *AppserverService {
	return &AppserverService{dbcPool: dbcPool, ctx: ctx}
}

func (s *AppserverService) PgTypeToPb(a *qx.Appserver) *pb_appserver.Appserver {
	return &pb_appserver.Appserver{
		Id:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppserverService) Create(name string, userId string) (*qx.Appserver, error) {
	// Keeping the validationErr variable as a way to show the pattern I'd like to follow (using a list of
	// validation errors to then send them)
	// Note: might change the pattern to use some sort of validation package. This might be duable by changing the
	// parameter in this method for example, to a struct type that can be validated. (Similar concept of python's
	// Pydantic object validation)
	validationErr := []string{}
	if name == "" {
		validationErr = AddValidationError("name", validationErr)
	}

	if userId == "" {
		validationErr = AddValidationError("user_id", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): missing name attribute", ValidationError))
	}

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): %v", ValidationError, err))
	}

	as, err := qx.New(s.dbcPool).CreateAppserver(s.ctx, qx.CreateAppserverParams{
		Name:      name,
		AppuserID: parsedUserId,
	})

	return &as, err
}

func (s *AppserverService) GetById(id string) (*qx.Appserver, error) {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	as, err := qx.New(s.dbcPool).GetAppserverById(s.ctx, parsedUuid)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &as, nil
}

func (s *AppserverService) List(name *wrappers.StringValue, ownerId string) ([]qx.Appserver, error) {
	// To query remember do to: {"name": {"value": "boo"}}
	var fName = pgtype.Text{Valid: false}

	if name != nil {
		fName.Valid = true
		fName.String = name.Value
	}

	parsedOwnerUuid, _ := uuid.Parse(ownerId)
	appservers, err := qx.New(s.dbcPool).ListUserAppservers(
		s.ctx, qx.ListUserAppserversParams{Name: fName, AppuserID: parsedOwnerUuid},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appservers, nil
}

func (s *AppserverService) Delete(id string, ownerId string) error {
	parsedUuid, err := uuid.Parse(id)

	if err != nil {
		return err
	}

	parsedOwnerUuid, _ := uuid.Parse(ownerId)

	deleted, err := qx.New(s.dbcPool).DeleteAppserver(
		s.ctx, qx.DeleteAppserverParams{ID: parsedUuid, AppuserID: parsedOwnerUuid})

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return err
}
