package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

func NewAppserverService(dbConn qx.DBTX, ctx context.Context) *AppserverService {
	return &AppserverService{dbConn: dbConn, ctx: ctx}
}

func (s *AppserverService) PgTypeToPb(a *qx.Appserver) *pb_appserver.Appserver {
	return &pb_appserver.Appserver{
		Id:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppserverService) Create(obj qx.CreateAppserverParams) (*qx.Appserver, error) {
	appserver, err := qx.New(s.dbConn).CreateAppserver(s.ctx, obj)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}
	
	return &appserver, err
}

func (s *AppserverService) GetById(id uuid.UUID) (*qx.Appserver, error) {
	appserver, err := qx.New(s.dbConn).GetAppserverById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, errors.New(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return &appserver, nil
}

func (s *AppserverService) List(name *wrappers.StringValue, ownerId string) ([]qx.Appserver, error) {
	// To query remember do to: {"name": {"value": "boo"}}
	var fName = pgtype.Text{Valid: false}

	if name != nil {
		fName.Valid = true
		fName.String = name.Value
	}

	parsedOwnerUuid, _ := uuid.Parse(ownerId)
	appservers, err := qx.New(s.dbConn).ListUserAppservers(
		s.ctx, qx.ListUserAppserversParams{Name: fName, AppuserID: parsedOwnerUuid},
	)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appservers, nil
}

func (s *AppserverService) Delete(obj qx.DeleteAppserverParams) error {
	deleted, err := qx.New(s.dbConn).DeleteAppserver(s.ctx, obj)

	if err != nil {
		return errors.New(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return errors.New(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return err
}
