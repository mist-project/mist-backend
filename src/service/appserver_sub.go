package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appserver"
	"mist/src/protos/v1/appserver_sub"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/qx"
)

type AppserverSubService struct {
	ctx  context.Context
	deps *ServiceDeps
}

func NewAppserverSubService(ctx context.Context, deps *ServiceDeps) *AppserverSubService {
	return &AppserverSubService{ctx: ctx, deps: deps}
}

func (s *AppserverSubService) PgTypeToPb(aSub *qx.AppserverSub) *appserver_sub.AppserverSub {
	return &appserver_sub.AppserverSub{
		Id:          aSub.ID.String(),
		AppserverId: aSub.AppserverID.String(),
		CreatedAt:   timestamppb.New(aSub.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aSub.UpdatedAt.Time),
	}
}

func (s *AppserverSubService) PgAppserverSubRowToPb(res *qx.ListUserServerSubsRow) *appserver_sub.AppserverAndSub {
	appserver := &appserver.Appserver{
		Id:        res.ID.String(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt.Time),
		UpdatedAt: timestamppb.New(res.UpdatedAt.Time),
	}

	return &appserver_sub.AppserverAndSub{
		Appserver: appserver,
		SubId:     res.AppserverSubID.String(),
	}
}

func (s *AppserverSubService) PgUserSubRowToPb(res *qx.ListAppserverUserSubsRow) *appserver_sub.AppuserAndSub {
	appuser := &appuser.Appuser{
		Id:        res.AppuserID.String(),
		Username:  res.AppuserUsername,
		CreatedAt: timestamppb.New(res.AppuserCreatedAt.Time),
		UpdatedAt: timestamppb.New(res.AppuserUpdatedAt.Time),
	}

	return &appserver_sub.AppuserAndSub{
		Appuser: appuser,
		SubId:   res.AppserverSubID.String(),
	}
}

// Creates a user to server subscription
func (s *AppserverSubService) Create(obj qx.CreateAppserverSubParams) (*qx.AppserverSub, error) {
	appserverSub, err := s.deps.Db.CreateAppserverSub(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &appserverSub, err
}

// Lists all the servers a user is subscribed to.
func (s *AppserverSubService) ListUserServerSubs(userId uuid.UUID) ([]qx.ListUserServerSubsRow, error) {
	/* Returns all servers a user belongs to. */

	subs, err := s.deps.Db.ListUserServerSubs(s.ctx, userId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return subs, nil
}

// Lists all the users in a server.
func (s *AppserverSubService) ListAppserverUserSubs(appserverId uuid.UUID) ([]qx.ListAppserverUserSubsRow, error) {

	subs, err := s.deps.Db.ListAppserverUserSubs(s.ctx, appserverId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return subs, nil
}

// Gets an appserver sub by its id.
func (s *AppserverSubService) GetById(id uuid.UUID) (*qx.AppserverSub, error) {
	role, err := s.deps.Db.GetAppserverSubById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError(fmt.Sprintf("unable to find appserver sub with id: %v", id), slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &role, nil
}

// Filters appserver subs.
func (s *AppserverSubService) Filter(args qx.FilterAppserverSubParams) ([]qx.FilterAppserverSubRow, error) {

	subs, err := s.deps.Db.FilterAppserverSub(s.ctx, args)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return subs, nil
}

// Removes user from server.
func (s *AppserverSubService) Delete(id uuid.UUID) error {
	// TODO: doing double queries here "fetching" the sub and then deleting it. maybe change this so that
	// we can do it in one query.
	sub, subErr := s.deps.Db.GetAppserverSubById(s.ctx, id)
	deleted, err := s.deps.Db.DeleteAppserverSub(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("unable to find appserver sub with id: %v", id), slog.LevelDebug)
	}

	if subErr == nil {
		user := []*appuser.Appuser{
			{Id: sub.AppuserID.String()},
		}

		s.deps.MProducer.SendMessage(
			context.Background(),
			os.Getenv("REDIS_NOTIFICATION_CHANNEL"),
			appserver.Appserver{Id: id.String()},
			event.ActionType_ACTION_REMOVE_SERVER, user,
		)
	}

	return nil
}
