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
	ctx context.Context, objId *string, action Action, subAction string,
) error {

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

	// get object and get permission role if exists
	if objId != nil {
		obj, err = GetObject(ctx, auth.shared, *objId, service.NewAppserverService(ctx, auth.DbConn, auth.Db).GetById)
		if err != nil {
			return err
		}
	}

	switch action {

	case ActionRead:
		switch subAction {
		case SubActionGetById:
			return nil
		case SubActionList:
			return nil
		}

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
