package permission

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mist/src/faults/message"
	"mist/src/middleware"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

type AppserverAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverAuthorizer {
	return &AppserverAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *AppserverAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action,
) error {

	if action == ActionCreate || action == ActionRead {
		// any user can create an appserver
		return nil
	}

	var (
		claims *middleware.CustomJWTClaims
		err    error
		obj    *qx.Appserver
		userId uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)

	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return message.ValidateError(message.InvalidUUID)
	}

	if objId == nil {
		// only on create we don't expect an object id
		return message.UnauthorizedError(message.Unauthorized)
	}

	obj, err = GetObject(ctx, auth.shared, objId, service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById)
	if err != nil {
		// if the object is not found or invalid uuid, we return error
		return message.UnauthorizedError(message.Unauthorized)

	}

	if obj.AppuserID == userId {
		return nil // user is the owner of the server, user can do anything
	}

	return message.UnauthorizedError(message.Unauthorized)
}
