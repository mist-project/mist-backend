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
		authctx    *AppserverIdAuthCtx
		authOk     bool
		claims     *middleware.CustomJWTClaims
		err        error
		obj        *qx.AppserverSub
		permission *qx.AppserverPermission
		userId     uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)
	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return message.ValidateError(message.InvalidUUID)
	}

	// get object and get permission role if exists
	if objId != nil {
		obj, err = GetObject(ctx, auth.shared, *objId, service.NewAppserverSubService(ctx, auth.DbConn, auth.Db, nil).GetById)
		if err != nil {
			return err
		}

		permission, _ = service.NewAppserverPermissionService(
			ctx, auth.DbConn, auth.shared.Db,
		).GetAppserverPermissionForUser(
			qx.GetAppserverPermissionForUserParams{AppserverID: obj.AppserverID, AppuserID: userId},
		)
	}

	authctx, authOk = ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx)

	// if permission role undefined, and auth context provided, attempt to get permission
	if authOk && permission == nil {
		permission, _ = service.NewAppserverPermissionService(
			ctx, auth.DbConn, auth.shared.Db,
		).GetAppserverPermissionForUser(
			qx.GetAppserverPermissionForUserParams{AppserverID: authctx.AppserverId, AppuserID: userId},
		)
	}

	switch action {
	case ActionRead:

		if permission != nil && permission.ReadAll.Bool {
			// user has elevated read permissions
			return nil
		}

		switch subAction {
		case SubActionListUserServerSubs:
			return nil
		case SubActionListAppserverUserSubs:
			return auth.canListAppserverSubs(ctx, userId, authctx)
		}

	case ActionWrite:

		if permission != nil && permission.WriteAll.Bool {
			// user has elevated write permissions
			return nil
		}

		switch subAction {
		case SubActionCreate:
			// Anyone can become a sub for a server.
			return nil
		}

	case ActionDelete:
		if permission != nil && permission.DeleteAll.Bool {
			// user has elevated delete permissions
			return nil
		}

		return auth.canDelete(ctx, userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// A user can only request list users subscribed to a server if they are subscribed to it.
func (auth *AppserverSubAuthorizer) canListAppserverSubs(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
	var (
		owner  bool
		hasSub bool
		err    error
	)

	if owner, err = auth.shared.UserIsServerOwner(ctx, userId, authCtx.AppserverId); owner {
		return nil
	}

	if hasSub, err = auth.shared.UserHasServerSub(ctx, userId, authCtx.AppserverId); err != nil {
		return err
	}

	if hasSub {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Server owner and and object owner can delete a subscription.
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
