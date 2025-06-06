package permission

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

type AppserverRoleSubAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewAppserverRoleSubAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *AppserverRoleSubAuthorizer {
	return &AppserverRoleSubAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *AppserverRoleSubAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action,
) error {
	var (
		authOk bool
		claims *middleware.CustomJWTClaims

		allowed     bool
		err         error
		permissions *PermissionMasks
		server      *qx.Appserver
		serverIdCtx *AppserverIdAuthCtx
		userId      uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )

	claims, _ = middleware.GetJWTClaims(ctx)

	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return faults.AuthorizationError(fmt.Sprintf("invalid user id: %v", err), slog.LevelDebug)
	}

	serverIdCtx, authOk = ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx)

	if !authOk {
		// if the object is not found or invalid uuid, we return error
		return faults.AuthorizationError(fmt.Sprintf("invalid %s in context", PermissionCtxKey), slog.LevelDebug)
	}

	allowed, err = auth.shared.BasePermissionCheck(ctx, serverIdCtx.AppserverId, userId, action)

	if err != nil {
		return faults.ExtendError(err)
	}

	if allowed {
		return nil // user has base permission, no need to check further
	}

	if objId != nil {
		_, err = GetObject(ctx, auth.shared, objId, service.NewAppserverRoleSubService(ctx, auth.DbConn, auth.Db, nil).GetById)

		if err != nil {
			// if the object is not found or invalid uuid, we return err
			return faults.ExtendError(err)
		}
	}

	server, err = service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById(serverIdCtx.AppserverId)

	if err != nil {
		// if the object is not found or invalid uuid, we return error
		return faults.ExtendError(err)
	}

	if server.AppuserID == userId {
		return nil // user is the owner of the server, user can do anything
	}

	permissions, err = GetUserPermissionMask(ctx, auth.shared, userId, server)

	if err != nil {
		return faults.ExtendError(err)
	}

	if permissions.AppserverPermissionMask&ManageRoles != 0 {
		return nil
	}

	return faults.AuthorizationError(fmt.Sprintf("user %s is not authorized to perform this action", userId), slog.LevelDebug)
}
