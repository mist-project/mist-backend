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
		return message.UnauthorizedError(message.Unauthorized)
	}

	serverIdCtx, authOk = ctx.Value(PermissionCtxKey).(*AppserverIdAuthCtx)

	if !authOk {
		// if the object is not found or invalid uuid, we return error
		return message.UnauthorizedError(message.Unauthorized)
	}

	allowed, err = auth.shared.BasePermissionCheck(ctx, serverIdCtx.AppserverId, userId, action)

	if err != nil {
		return message.UnauthorizedError(message.Unauthorized)
	}

	if allowed {
		return nil // user has base permission, no need to check further
	}

	if objId != nil {
		_, err = GetObject(ctx, auth.shared, objId, service.NewAppserverRoleService(ctx, auth.DbConn, auth.Db).GetById)

		if err != nil {
			// if the object is not found or invalid uuid, we return err
			return err
		}
	}

	server, err = service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById(serverIdCtx.AppserverId)

	if err != nil {
		// if the object is not found or invalid uuid, we return error
		return message.UnauthorizedError(message.Unauthorized)
	}

	if server.AppuserID == userId {
		return nil // user is the owner of the server, user can do anything
	}

	permissions, err = GetUserPermissionMask(ctx, auth.shared, userId, server)

	if err != nil {
		return message.UnauthorizedError(message.Unauthorized)
	}

	if permissions.AppserverPermissionMask&ManageRoles != 0 {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}
