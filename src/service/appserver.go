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
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	ctx  context.Context
	deps *ServiceDeps
}

// Creates a new AppserverService struct.
func NewAppserverService(
	ctx context.Context, deps *ServiceDeps) *AppserverService {
	return &AppserverService{ctx: ctx, deps: deps}
}

// Converts a database appserver object to protobuff appserver object
func (s *AppserverService) PgTypeToPb(a *qx.Appserver) *appserver.Appserver {
	return &appserver.Appserver{
		Id:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

// Creates an appserver, uses CreateWithTx helper function to wrap creation in transaction
// Note: the transaction will be committed in CreateWithTx. The creator of the server gets automatically assigned
// an appserver sub.
func (s *AppserverService) Create(obj qx.CreateAppserverParams) (*qx.Appserver, error) {
	appserver, err := s.deps.Db.CreateAppserver(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("create appserver error: %v", err), slog.LevelError)
	}

	// once the appserver is created, add user as a subscriber
	_, err = s.deps.Db.CreateAppserverSub(
		s.ctx,
		qx.CreateAppserverSubParams{AppserverID: appserver.ID, AppuserID: obj.AppuserID},
	)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("create appserver sub error: %v", err), slog.LevelError)
	}

	// if err := tx.Commit(s.ctx); err != nil {
	// 	return nil, faults.DatabaseError(fmt.Sprintf("database error commit: %v", err), slog.LevelError)
	// }

	return &appserver, err
}

// Gets an appserver detail by its id.
func (s *AppserverService) GetById(id uuid.UUID) (*qx.Appserver, error) {
	appserver, err := s.deps.Db.GetAppserverById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError(fmt.Sprintf("unable to find appserver with id: %v", id), slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &appserver, nil
}

// Lists all appservers based on the owner. Name filter is also added but it may get deprecated.
func (s *AppserverService) List(params qx.ListAppserversParams) ([]qx.Appserver, error) {
	appservers, err := s.deps.Db.ListAppservers(s.ctx, params)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return appservers, nil
}

// Delete appserver object, for now only owners can delete an appserver.
func (s *AppserverService) Delete(id uuid.UUID) error {

	// Get all subs for the appserver
	subs, err := s.deps.Db.ListAppserverUserSubs(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelWarn)
	}

	deleted, err := s.deps.Db.DeleteAppserver(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("unable to find appserver with id: %v", id), slog.LevelDebug)
	}

	if len(subs) > 0 {
		s.SendDeleteNotificationToUsers(subs, id)
	}

	return err
}

func (s *AppserverService) SendDeleteNotificationToUsers(subs []qx.ListAppserverUserSubsRow, appserverID uuid.UUID) {

	users := make([]*appuser.Appuser, 0, len(subs))

	for _, sub := range subs {
		users = append(users, &appuser.Appuser{
			Id:       sub.AppuserID.String(),
			Username: sub.AppuserUsername,
		})
	}

	s.deps.MProducer.SendMessage(
		context.Background(),
		os.Getenv("REDIS_NOTIFICATION_CHANNEL"),
		&appserver.Appserver{Id: appserverID.String()},
		event.ActionType_ACTION_REMOVE_SERVER, users,
	)
}
