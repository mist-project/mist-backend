package rpcs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *ChannelGRPCService) Create(ctx context.Context, req *channel.CreateRequest) (*channel.CreateResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionCreate); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	cs := service.NewChannelService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	c, err := cs.Create(qx.CreateChannelParams{Name: req.Name, AppserverID: serverId, IsPrivate: req.IsPrivate})

	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	return &channel.CreateResponse{
		Channel: cs.PgTypeToPb(c),
	}, nil

}

func (s *ChannelGRPCService) GetById(
	ctx context.Context, req *channel.GetByIdRequest,
) (*channel.GetByIdResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	cs := service.NewChannelService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	id, _ := uuid.Parse(req.Id)
	c, err := cs.GetById(id)

	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	return &channel.GetByIdResponse{Channel: cs.PgTypeToPb(c)}, nil
}

func (s *ChannelGRPCService) ListServerChannels(
	ctx context.Context, req *channel.ListServerChannelsRequest,
) (*channel.ListServerChannelsResponse, error) {

	var (
		err        error
		nameFilter pgtype.Text
	)
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	cs := service.NewChannelService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)

	if req.Name != nil {
		nameFilter = pgtype.Text{Valid: true, String: req.Name.Value}
	}

	channels, _ := cs.ListServerChannels(qx.ListServerChannelsParams{Name: nameFilter, AppserverID: serverId})
	response := &channel.ListServerChannelsResponse{}
	response.Channels = make([]*channel.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, cs.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *ChannelGRPCService) Delete(
	ctx context.Context, req *channel.DeleteRequest,
) (*channel.DeleteResponse, error) {

	var err error
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	id, _ := uuid.Parse(req.Id)
	if err := service.NewChannelService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).Delete(id); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	return &channel.DeleteResponse{}, nil
}
