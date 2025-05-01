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

type AppserverRoleAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverRoleAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverRoleAuthorizer {
	return &AppserverRoleAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *AppserverRoleAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action, subAction string,
) error {

	var (
		authctx    *AppserverIdAuthCtx
		authOk     bool
		claims     *middleware.CustomJWTClaims
		err        error
		obj        *qx.AppserverRole
		permission *qx.AppserverPermission
		userId     uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)
	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return message.ValidateError(message.InvalidUUID)
	}

	if objId != nil {
		obj, err = GetObject(ctx, auth.shared, *objId, service.NewAppserverRoleService(ctx, auth.DbConn, auth.Db).GetById)
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
		case SubActionListServerRoles:
			return auth.canListServerRoles(ctx, userId, authctx)
		}

	case ActionWrite:

		if permission != nil && permission.WriteAll.Bool {
			// user has elevated read permissions
			return nil
		}

		switch subAction {
		case SubActionCreate:
			return auth.canCreate(ctx, userId, authctx)
		}

	case ActionDelete:

		if permission != nil && permission.DeleteAll.Bool {
			// user has elevated read permissions
			return nil
		}

		return auth.canDelete(ctx, userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// A user can only request all channels in a server if they are subscribed to it.
func (auth *AppserverRoleAuthorizer) canListServerRoles(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
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

// Only server owners can create roles.
func (auth *AppserverRoleAuthorizer) canCreate(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
	var (
		owner bool
		err   error
	)

	if owner, err = auth.shared.UserIsServerOwner(ctx, userId, authCtx.AppserverId); err != nil {
		return err
	}

	if owner {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Only server owners can delete roles.
func (auth *AppserverRoleAuthorizer) canDelete(ctx context.Context, userId uuid.UUID, obj *qx.AppserverRole) error {
	var (
		owner bool
		err   error
	)

	if owner, err = auth.shared.UserIsServerOwner(ctx, userId, obj.AppserverID); err != nil {
		return err
	}

	if owner {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}
