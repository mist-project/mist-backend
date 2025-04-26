package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	dbConn qx.DBTX
	ctx    context.Context
	db     db.Querier
}

func NewAppserverSubService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverSubService {
	return &AppserverSubService{ctx: ctx, dbConn: dbConn, db: db}
}

func TempNewAppserverSubService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverSubService {
	return &AppserverSubService{dbConn: dbConn, ctx: ctx, db: db}
}

func (s *AppserverSubService) PgTypeToPb(aSub *qx.AppserverSub) *pb_appserversub.AppserverSub {
	return &pb_appserversub.AppserverSub{
		Id:          aSub.ID.String(),
		AppserverId: aSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(aSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aSub.UpdatedAt.Time),
	}
}

func (s *AppserverSubService) PgAppserverSubRowToPb(res *qx.GetUserAppserverSubsRow) *pb_appserversub.AppserverAndSub {
	appserver := &pb_appserver.Appserver{
		Id:        res.ID.String(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserversub.AppserverAndSub{
		Appserver: appserver,
		SubId:     res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) PgUserSubRowToPb(res *qx.GetAllUsersAppserverSubsRow) *pb_appserversub.AppuserAndSub {
	appuser := &pb_appuser.Appuser{
		Id:        res.ID.String(),
		Username:  res.Username,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserversub.AppuserAndSub{
		Appuser: appuser,
		SubId:   res.AppserverSubID.String(),
	}
}

// Creates a user to server subscription
func (s *AppserverSubService) Create(obj qx.CreateAppserverSubParams) (*qx.AppserverSub, error) {
	appserverSub, err := s.db.CreateAppserverSub(s.ctx, obj)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return &appserverSub, err
}

// Creates a user to server subscription using injected transaction, does not commit the transaction.
func (s *AppserverSubService) CreateWithTx(obj qx.CreateAppserverSubParams, tx pgx.Tx) (*qx.AppserverSub, error) {
	txQ := s.db.WithTx(tx)
	appserverSub, err := txQ.CreateAppserverSub(s.ctx, obj)

	return &appserverSub, err
}

// Lists all the servers a user is subscribed to.
func (s *AppserverSubService) ListUserAppserverAndSub(userId uuid.UUID) ([]qx.GetUserAppserverSubsRow, error) {
	/* Returns all servers a user belongs to. */

	subs, err := s.db.GetUserAppserverSubs(s.ctx, userId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return subs, nil
}

// Lists all the users in a server.
func (s *AppserverSubService) ListAllUsersAppserverAndSub(
	appserverId uuid.UUID,
) ([]qx.GetAllUsersAppserverSubsRow, error) {

	subs, err := s.db.GetAllUsersAppserverSubs(s.ctx, appserverId)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return subs, nil
}

// Removes user from server.
func (s *AppserverSubService) DeleteByAppserver(id uuid.UUID) error {
	deleted, err := s.db.DeleteAppserverSub(s.ctx, id)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
	}

	return nil
}
