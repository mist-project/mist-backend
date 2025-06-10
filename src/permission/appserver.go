package permission

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

type AppserverAuthorizer struct {
	DbTx   pgx.Tx
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverAuthorizer(Db db.Querier) *AppserverAuthorizer {
	return &AppserverAuthorizer{
		Db: Db,
		shared: &SharedAuthorizer{
			Db: Db,
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
		return faults.AuthorizationError(fmt.Sprintf("invalid user id: %s", claims.UserID), slog.LevelDebug)
	}

	if objId == nil {
		// only on create we don't expect an object id
		return faults.AuthorizationError(fmt.Sprintf("object id is required for action: %s", action), slog.LevelDebug)
	}

	obj, err = GetObject(ctx, auth.shared, objId, service.NewAppserverService(ctx, &service.ServiceDeps{Db: auth.Db}).GetById)
	if err != nil {
		// if the object is not found or invalid uuid, we return error
		return faults.ExtendError(err)
	}

	if obj.AppuserID == userId {
		return nil // user is the owner of the server, user can do anything
	}

	return faults.AuthorizationError("user is not allowed to manage server", slog.LevelDebug)
}
