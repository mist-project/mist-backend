package permission

import (
	"context"
	"mist/src/errors/message"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SharedAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
}

type AppserverIdAuthCtx struct {
	AppserverId uuid.UUID
}

func NewSharedAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *SharedAuthorizer {
	return &SharedAuthorizer{
		DbConn: DbConn,
		Db:     Db,
	}
}

// Helper function to determine whether a user is owner of the server.
func (auth *SharedAuthorizer) UserIsServerOwner(ctx context.Context, userId uuid.UUID, serverId uuid.UUID) (bool, error) {
	server, err := service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById(serverId)

	if err != nil {
		return false, err
	}

	return server.AppuserID == userId, nil
}

// Helper function to determine whether a user is owner of the server.
func (auth *SharedAuthorizer) UserHasServerSub(ctx context.Context, userId uuid.UUID, serverId uuid.UUID) (bool, error) {
	sub, err := service.NewAppserverSubService(ctx, auth.DbConn, auth.Db, nil).Filter(
		qx.FilterAppserverSubParams{
			AppserverID: pgtype.UUID{Valid: true, Bytes: serverId},
			AppuserID:   pgtype.UUID{Valid: true, Bytes: userId},
		},
	)

	if err != nil {
		return false, err
	} else if len(sub) > 0 {
		return true, nil
	}

	return false, nil
}

func GetObject[T any](
	ctx context.Context, auth *SharedAuthorizer, objId string, fetchFunc func(uuid.UUID) (*T, error),
) (*T, error) {

	id, err := uuid.Parse(objId)
	if err != nil {
		return nil, message.ValidateError(message.InvalidUUID)
	}

	obj, err := fetchFunc(id)

	if err != nil {
		return nil, message.NotFoundError(message.NotFound)
	}

	return obj, nil
}
