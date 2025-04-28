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

type AppserverSubAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverSubAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverSubAuthorizer {
	return &AppserverSubAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *AppserverSubAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action, subAction string,
) error {

	var (
		err    error
		obj    *qx.AppserverSub
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
		svc := service.NewAppserverSubService(ctx, auth.DbConn, auth.Db)
		obj, err = svc.GetById(id)

		if err != nil {
			return message.NotFoundError(message.NotFound)
		}
	}
	// ---------------------

	switch action {
	case ActionRead:
		switch subAction {
		case SubActionListUserServerSubs:
			return nil
		case SubActionListAppserverUserSubs:
			return auth.canListAppserverSubs(ctx, userId, ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx))
		}

	case ActionWrite:
		switch subAction {
		case SubActionCreate:
			// Anyone can become a sub for a server.
			return nil
		}
	case ActionDelete:
		return auth.canDelete(ctx, userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// A user can only request list users subscribed to a server if they are subscribed to it.
func (auth *AppserverSubAuthorizer) canListAppserverSubs(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
	var (
		hasSub bool
		err    error
	)

	if hasSub, err = auth.shared.UserHasServerSub(ctx, userId, authCtx.AppserverId); err != nil {
		return err
	}

	if hasSub {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Server owner and and object owner can delete a subscription.
// TODO: with permissions allow other users to delete subs (pending ServerPermission definition)
func (auth *AppserverSubAuthorizer) canDelete(ctx context.Context, userId uuid.UUID, obj *qx.AppserverSub) error {
	var (
		owner bool
		err   error
	)

	if userId == obj.AppuserID {
		return nil
	}

	if owner, err = auth.shared.UserIsServerOwner(ctx, userId, obj.AppserverID); err != nil {
		return err
	}

	if owner {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}
