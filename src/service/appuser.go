package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	dbcPool *pgxpool.Pool
	ctx     context.Context
}

func NewAppuserService(dbcPool *pgxpool.Pool, ctx context.Context) *AppuserService {
	return &AppuserService{dbcPool: dbcPool, ctx: ctx}
}

func (s *AppuserService) PgTypeToPb(a *qx.Appuser) *pb_appuser.Appuser {
	return &pb_appuser.Appuser{
		Id:        a.ID.String(),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppuserService) Create(name string, userId string) (*qx.Appuser, error) {
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
		validationErr = AddValidationError("id", validationErr)
	}

	if len(validationErr) > 0 {
		return nil, errors.New(fmt.Sprintf("(%d): missing name attribute", ValidationError))
	}

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): %v", ValidationError, err))
	}

	as, err := qx.New(s.dbcPool).CreateAppuser(s.ctx, qx.CreateAppuserParams{
		ID:       parsedUserId,
		Username: name,
	})

	return &as, err
}
