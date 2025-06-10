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

type ChannelAuthorizer struct {
	DbTx   pgx.Tx
	Db     db.Querier
	shared *SharedAuthorizer
}

func NewChannelAuthorizer(Db db.Querier) *ChannelAuthorizer {
	return &ChannelAuthorizer{
		Db: Db,
		shared: &SharedAuthorizer{
			Db: Db,
		},
	}
}

func (auth *ChannelAuthorizer) Authorize(
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
		return faults.AuthorizationError(fmt.Sprintf("invalid user id: %s", claims.UserID), slog.LevelDebug)
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
		_, err = GetObject(ctx, auth.shared, objId, service.NewChannelService(ctx, &service.ServiceDeps{Db: auth.Db}).GetById)

		if err != nil {
			// if the object is not found or invalid uuid, we return error
			return faults.ExtendError(err)
		}
	}

	server, err = service.NewAppserverService(ctx, &service.ServiceDeps{Db: auth.Db}).GetById(serverIdCtx.AppserverId)

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

	if permissions.AppserverPermissionMask&ManageChannels != 0 {
		return nil
	}

	return faults.AuthorizationError("user does not have permission to manage channels", slog.LevelDebug)
}
