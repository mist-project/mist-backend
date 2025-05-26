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

type AppserverPermissionAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverPermissionAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverPermissionAuthorizer {
	return &AppserverPermissionAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *AppserverPermissionAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action, subAction string,
) error {

	var (
		err    error
		obj    *qx.AppserverPermission
		claims *middleware.CustomJWTClaims
		userId uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)
	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return message.ValidateError(message.InvalidUUID)
	}

	if objId != nil {
		obj, err = GetObject(ctx, auth.shared, objId, service.NewAppserverPermissionService(ctx, auth.DbConn, auth.Db).GetById)
		if err != nil {
			return err
		}
	}

	authctx, _ := ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx)

	switch action {
	case ActionRead:
		switch subAction {
		case SubActionListAppserverUserPermsission:

			return auth.canListServerUserPermission(ctx, userId, authctx)
		}
	case ActionWrite:
		switch subAction {
		case SubActionCreate:
			return auth.canCreate(ctx, userId, authctx)
		}
	case ActionDelete:
		return auth.canDelete(ctx, userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Only server owners can retreive all appserver permission role users.
func (auth *AppserverPermissionAuthorizer) canListServerUserPermission(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
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

// Only server owners can create appserver permission roles.
func (auth *AppserverPermissionAuthorizer) canCreate(ctx context.Context, userId uuid.UUID, authCtx *AppserverIdAuthCtx) error {
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

// Only server owners can delete appserver permission roles.
func (auth *AppserverPermissionAuthorizer) canDelete(ctx context.Context, userId uuid.UUID, obj *qx.AppserverPermission) error {
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
