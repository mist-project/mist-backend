package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverService {
	return &AppserverService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverService) PgTypeToPb(a *qx.Appserver) *pb_appserver.Appserver {
	return &pb_appserver.Appserver{
		Id:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppserverService) Create(obj qx.CreateAppserverParams) (*qx.Appserver, error) {
	tx, err := s.dbConn.BeginTx(s.ctx, pgx.TxOptions{})
	fmt.Printf("boom")
	if err != nil {
		return nil, fmt.Errorf("(%d): failed to start transaction- %v", DatabaseError, err)
	}
	defer tx.Rollback(s.ctx)
	txQ := s.db.WithTx(tx)

	appserver, err := txQ.CreateAppserver(s.ctx, obj)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error- %v", DatabaseError, err))
	}

	// once the appserver is created, add user as a subscriber
	_, err = NewAppserverSubService(tx, s.ctx).Create(
		qx.CreateAppserverSubParams{
			AppserverID: appserver.ID,
			AppuserID:   obj.AppuserID,
		},
	)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error- %v", DatabaseError, err))
	}

	if err := tx.Commit(s.ctx); err != nil {
		return nil, fmt.Errorf("(%d): database error- %v", DatabaseError, err)
	}

	return &appserver, err
}

func (s *AppserverService) GetById(id uuid.UUID) (*qx.Appserver, error) {
	appserver, err := s.db.GetAppserverById(s.ctx, id)

	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf(fmt.Sprintf("(%d): resource not found", NotFoundError))
		}

		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
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
	appservers, err := s.db.ListUserAppservers(
		s.ctx, qx.ListUserAppserversParams{Name: fName, AppuserID: parsedOwnerUuid},
	)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	}

	return appservers, nil
}

func (s *AppserverService) Delete(obj qx.DeleteAppserverParams) error {
	deleted, err := s.db.DeleteAppserver(s.ctx, obj)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d): database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d): no rows were deleted", NotFoundError))
	}

	return err
}
