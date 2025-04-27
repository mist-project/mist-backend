package permission

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mist/src/errors/message"
	"mist/src/middleware"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

type AppserverAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
}

func NewAppserverAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverAuthorizer {
	return &AppserverAuthorizer{
		DbConn: DbConn,
		Db:     Db,
	}
}

func (auth *AppserverAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action, subAction string,
) error {

	var (
		err    error
		obj    *qx.Appserver
		claims *middleware.CustomJWTClaims
		userId uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)
	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return message.ValidateError(message.InvalidUUID)
	}

	// ---- GET OBJECT -----
	// TODO: refactor this to potentially generalize
	if objId != nil {
		// Get object if id provided
		id, err := uuid.Parse(*objId)
		if err != nil {
			return message.ValidateError(message.InvalidUUID)
		}

		svc := service.NewAppserverService(ctx, auth.DbConn, auth.Db)
		obj, err = svc.GetById(id)

		if err != nil {
			return message.NotFoundError(message.NotFound)
		}
	}
	// ---------------------

	switch action {
	case ActionRead:
		switch subAction {
		case SubActionGetById:
			return nil
		}
		return nil
	case ActionWrite:
		switch subAction {
		case SubActionCreate:
			return nil
		}
	case ActionDelete:
		return auth.canDelete(userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Only server owners can delete a server.
func (auth *AppserverAuthorizer) canDelete(userId uuid.UUID, obj *qx.Appserver) error {
	if userId == obj.AppuserID {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}
