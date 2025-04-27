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

type ChannelAuthorizer struct {
	DbConn *pgxpool.Pool
	Db     db.Querier
	shared *SharedAuthorizer
}

type ChannelListAppserverChannelCtx struct {
	AppserverId uuid.UUID
}

type ChannelCreateCtx struct {
	AppserverId uuid.UUID
}

func NewChannelAuthorizer(DbConn *pgxpool.Pool, Db db.Querier) *ChannelAuthorizer {
	return &ChannelAuthorizer{
		DbConn: DbConn,
		Db:     Db,
		shared: &SharedAuthorizer{
			DbConn: DbConn,
			Db:     Db,
		},
	}
}

func (auth *ChannelAuthorizer) Authorize(
	ctx context.Context, objId *string, action Action, subAction string,
) error {

	var (
		err    error
		obj    *qx.Channel
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

		svc := service.NewChannelService(ctx, auth.DbConn, auth.Db)
		obj, err = svc.GetById(id)

		if err != nil {
			return message.NotFoundError(message.NotFound)
		}
	}
	// ---------------------

	switch action {
	case ActionRead:
		switch subAction {
		case SubActionListAppserverChannels:
			return auth.canListAppserverChannels(ctx, userId, ctx.Value(PermissionCtxKey).(*ChannelListAppserverChannelCtx))
		case SubActionGetById:
			return auth.canGetById(ctx, userId, obj)
		}
	case ActionWrite:
		switch subAction {
		case SubActionCreate:
			return auth.canCreate(ctx, userId, ctx.Value(PermissionCtxKey).(*ChannelCreateCtx))
		}
	case ActionDelete:
		return auth.canDelete(ctx, userId, obj)
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// A user can only request all channels in a server if they are subscribed to it.
func (auth *ChannelAuthorizer) canListAppserverChannels(ctx context.Context, userId uuid.UUID, authCtx *ChannelListAppserverChannelCtx) error {
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

// A user can only request channel's details if they are subscribed to it
func (auth *ChannelAuthorizer) canGetById(ctx context.Context, userId uuid.UUID, channel *qx.Channel) error {
	var (
		hasSub bool
		err    error
	)

	if hasSub, err = auth.shared.UserHasServerSub(ctx, userId, channel.AppserverID); err != nil {
		return err
	}

	if hasSub {
		return nil
	}

	return message.UnauthorizedError(message.Unauthorized)
}

// Only server owners can create channels.
// TODO: with permissions allow other users to create channels (pending ServerPermission definition)
func (auth *ChannelAuthorizer) canCreate(ctx context.Context, userId uuid.UUID, authCtx *ChannelCreateCtx) error {
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

// Only server owners can delete channels.
// TODO: with permissions allow other users to delete channels (pending ServerPermission definition)
func (auth *ChannelAuthorizer) canDelete(ctx context.Context, userId uuid.UUID, obj *qx.Channel) error {
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
