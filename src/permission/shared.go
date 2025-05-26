package permission

import (
	"context"
	"mist/src/errors/message"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SharedAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
}

type AppserverIdAuthCtx struct {
	AppserverId uuid.UUID
}

type PermissionMasks struct {
	AppserverPermissionMask int64
	ChannelPermissionMask   int64
	SubPermissionMask       int64
}

func NewSharedAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *SharedAuthorizer {
	return &SharedAuthorizer{
		DbConn: DbConn,
		Db:     Db,
	}
}

// Helper function to determine whether a user is owner of the server.
func (auth *SharedAuthorizer) UserIsServerOwner(ctx context.Context, userId uuid.UUID, serverId uuid.UUID) (bool, error) {
	server, err := service.NewAppserverService(ctx, auth.DbConn, auth.Db, nil).GetById(serverId)

	if err != nil {
		return false, err
	}

	return server.AppuserID == userId, nil
}

// Helper function to determine whether a user is owner of the server.
func (auth *SharedAuthorizer) UserHasServerSub(ctx context.Context, userId uuid.UUID, serverId uuid.UUID) (bool, error) {
	sub, err := service.NewAppserverSubService(ctx, auth.DbConn, auth.Db, nil).Filter(
		qx.FilterAppserverSubParams{
			AppserverID: pgtype.UUID{Valid: true, Bytes: serverId},
			AppuserID:   pgtype.UUID{Valid: true, Bytes: userId},
		},
	)
	if err != nil {
		return false, err
	} else if len(sub) > 0 {
		return true, nil
	}

	return false, nil
}

func (auth *SharedAuthorizer) BasePermissionCheck(
	ctx context.Context, appserverId uuid.UUID, userId uuid.UUID, action Action,
) (bool, error) {
	var (
		err    error
		hasSub bool
	)

	if hasSub, err = auth.UserHasServerSub(ctx, userId, appserverId); err != nil {
		return false, message.UnauthorizedError(message.Unauthorized)
	}

	if hasSub && action == ActionRead {
		// if the user has a sub for this server, he can read it
		return true, nil
	}

	return false, nil
}

func GetObject[T any](
	ctx context.Context, auth *SharedAuthorizer, objId *string, fetchFunc func(uuid.UUID) (*T, error),
) (*T, error) {

	if objId == nil {
		return nil, message.ValidateError(message.InvalidUUID)
	}

	id, err := uuid.Parse(*objId)
	if err != nil {
		return nil, message.ValidateError(message.InvalidUUID)
	}

	obj, err := fetchFunc(id)

	if err != nil {
		return nil, message.NotFoundError(message.NotFound)
	}

	return obj, nil
}

func GetUserPermissionMask(
	ctx context.Context, auth *SharedAuthorizer, userId uuid.UUID, obj *qx.Appserver,
) (*PermissionMasks, error) {

	roles, err := service.NewAppserverRoleSubService(ctx, auth.DbConn, auth.Db).GetAppuserRoles(qx.GetAppuserRolesParams{
		AppserverID: obj.ID,
		AppuserID:   userId,
	})

	if err != nil {
		return nil, err
	}

	masks := PermissionMasks{
		AppserverPermissionMask: 0,
		ChannelPermissionMask:   0,
		SubPermissionMask:       0,
	}

	for _, role := range roles {
		masks.AppserverPermissionMask = role.AppserverPermissionMask | masks.AppserverPermissionMask
		masks.ChannelPermissionMask = role.ChannelPermissionMask | masks.ChannelPermissionMask
		masks.SubPermissionMask = role.SubPermissionMask | masks.SubPermissionMask
	}

	return &masks, nil
}
