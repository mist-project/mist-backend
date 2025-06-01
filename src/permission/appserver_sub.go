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
	ctx context.Context, objId *string, action Action,
) error {

	if action == ActionCreate {
		// any user can create an appserver sub
		return nil
	}

	var (
		authOk      bool
		claims      *middleware.CustomJWTClaims
		allowed     bool
		err         error
		server      *qx.Appserver
		serverIdCtx *AppserverIdAuthCtx
		sub         *qx.AppserverSub
		permissions *PermissionMasks
		userId      uuid.UUID
	)

	// No error expected when getting claims. this method should be hit AFTER authentication ( which sets claims )
	claims, _ = middleware.GetJWTClaims(ctx)

	if userId, err = uuid.Parse(claims.UserID); err != nil {
		return faults.AuthorizationError(fmt.Sprintf("invalid user id: %s", claims.UserID), slog.LevelDebug)
	}

	serverIdCtx, authOk = ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx)

	if !authOk {
		// if the object is not found or invalid uuid, we return error
		return faults.AuthorizationError(fmt.Sprintf("invalid %s in context", PermissionCtxKey), slog.LevelDebug)
	}

	allowed, err = auth.shared.BasePermissionCheck(ctx, serverIdCtx.AppserverId, userId, action)

	if err != nil {
		err, ok := faults.ExtendError(err).(*faults.CustomError)
		if ok {
			return faults.AuthorizationError(err.StackTrace(), slog.LevelDebug)
		}
	}

	if allowed {
		return nil // user has base permission, no need to check further
	}

	sub, err = GetObject(ctx, auth.shared, objId, service.NewAppserverSubService(ctx, auth.DbConn, auth.Db, nil).GetById)

	if err != nil {
		// if the object is not found or invalid uuid, we return error
		return faults.ExtendError(err)
	}

	server, err = service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById(serverIdCtx.AppserverId)

	if err != nil {
		// if the object is not found or invalid uuid, we return error
		err, ok := faults.ExtendError(err).(*faults.CustomError)
		if ok {
			return faults.AuthorizationError(fmt.Sprintf("failed to get appserver: %v", err.StackTrace()), slog.LevelDebug)
		}
	}

	if action == ActionDelete {

		if server.AppuserID == sub.AppuserID {
			// nobody can delete the owner's sub
			return faults.AuthorizationError("cannot delete the owner's sub", slog.LevelDebug)
		} else if sub.AppuserID == userId {
			// user can delete their own sub
			return nil
		}
	}

	if server.AppuserID == userId {
		return nil // user is the owner of the server, user can do anything
	}

	permissions, err = GetUserPermissionMask(ctx, auth.shared, userId, server)

	if err != nil {
		err, ok := faults.ExtendError(err).(*faults.CustomError)
		if ok {
			return faults.AuthorizationError(fmt.Sprintf("failed to get user permissions: %v", err.StackTrace()), slog.LevelDebug)
		}
	}

	if permissions.SubPermissionMask&ManageSubs != 0 {
		return nil
	}

	return faults.AuthorizationError("user does not have permission to manage subscriptions", slog.LevelDebug)
}
