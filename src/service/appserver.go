package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mist/src/psql_db/qx"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_mistbe "mist/src/protos/mistbe/v1"
)

type AppserverService struct {
	dbc_pool *pgxpool.Pool
	ctx      context.Context
}

func NewAppserverService(dbc_pool *pgxpool.Pool, ctx context.Context) *AppserverService {
	return &AppserverService{dbc_pool: dbc_pool, ctx: ctx}
}

func (service *AppserverService) PgTypeToPb(appserver *qx.Appserver) *pb_mistbe.Appserver {
	return &pb_mistbe.Appserver{
		Id:        appserver.ID.String(),
		Name:      appserver.Name,
		CreatedAt: timestamppb.New(appserver.CreatedAt.Time),
	}
}

func (service *AppserverService) Create(name string) (*qx.Appserver, error) {
	validation_errors := []string{}
	if name == "" {
		validation_errors = append(validation_errors, fmt.Sprintf("(%d): missing name attribute", ValidationError))
	}

	if len(validation_errors) > 0 {
		return nil, errors.New(strings.Join(validation_errors, "\n"))
	}

	appserver, err := qx.New(service.dbc_pool).CreateAppserver(service.ctx, name)
	return &appserver, err
}

func (service *AppserverService) GetById(id string) (qx.Appserver, error) {
	parsed_uuid, err := uuid.Parse(id)
	if err != nil {
		log.Fatalf("Invalid UUID string: %v", err)
	}
	return qx.New(service.dbc_pool).GetAppserver(service.ctx, parsed_uuid)
}

func (service *AppserverService) List(name *wrappers.StringValue) ([]qx.Appserver, error) {
	// To query remember do to: {"name": {"value": "boo"}}
	var formatName = pgtype.Text{Valid: false}
	if name != nil {
		formatName.Valid = true
		formatName.String = name.Value
	}
	return qx.New(service.dbc_pool).ListAppservers(service.ctx, formatName)
}

func (service *AppserverService) Delete(id string) error {
	parsed_uuid, err := uuid.Parse(id)

	deleted_rows, err := qx.New(service.dbc_pool).DeleteAppserver(service.ctx, parsed_uuid)
	if err != nil {
		log.Fatalf("error deleting: %v", err)
	} else if deleted_rows == 0 {
		log.Printf("No rows were deleted.")
	}
	return err
}
