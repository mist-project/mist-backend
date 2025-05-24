package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserver_sub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverSubService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverSubService {
	return &AppserverSubService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverSubService) PgTypeToPb(aSub *qx.AppserverSub) *pb_appserver_sub.AppserverSub {
	return &pb_appserver_sub.AppserverSub{
		Id:          aSub.ID.String(),
		AppserverId: aSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(aSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aSub.UpdatedAt.Time),
	}
}

func (s *AppserverSubService) PgAppserverSubRowToPb(res *qx.ListUserServerSubsRow) *pb_appserver_sub.AppserverAndSub {
	appserver := &pb_appserver.Appserver{
		Id:        res.ID.String(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserver_sub.AppserverAndSub{
		Appserver: appserver,
		SubId:     res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) PgUserSubRowToPb(res *qx.ListAppserverUserSubsRow) *pb_appserver_sub.AppuserAndSub {
	appuser := &pb_appuser.Appuser{
		Id:        res.ID.String(),
		Username:  res.Username,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &pb_appserver_sub.AppuserAndSub{
		Appuser: appuser,
		SubId:   res.AppserverSubID.String(),
	}
}

// Creates a user to server subscription
func (s *AppserverSubService) Create(obj qx.CreateAppserverSubParams) (*qx.AppserverSub, error) {
	appserverSub, err := s.db.CreateAppserverSub(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
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
func (s *AppserverSubService) ListUserServerSubs(userId uuid.UUID) ([]qx.ListUserServerSubsRow, error) {
	/* Returns all servers a user belongs to. */

	subs, err := s.db.ListUserServerSubs(s.ctx, userId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return subs, nil
}

// Lists all the users in a server.
func (s *AppserverSubService) ListAppserverUserSubs(appserverId uuid.UUID) ([]qx.ListAppserverUserSubsRow, error) {

	subs, err := s.db.ListAppserverUserSubs(s.ctx, appserverId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return subs, nil
}

// Gets an appserver sub by its id.
func (s *AppserverSubService) GetById(id uuid.UUID) (*qx.AppserverSub, error) {
	role, err := s.db.GetAppserverSubById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, message.NotFoundError(message.NotFound)
		}

		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &role, nil
}

// Filters appserver subs.
func (s *AppserverSubService) Filter(args qx.FilterAppserverSubParams) ([]qx.FilterAppserverSubRow, error) {

	subs, err := s.db.FilterAppserverSub(s.ctx, args)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return subs, nil
}

// Removes user from server.
func (s *AppserverSubService) Delete(id uuid.UUID) error {
	deleted, err := s.db.DeleteAppserverSub(s.ctx, id)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}

	return nil
}
