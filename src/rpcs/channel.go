package rpcs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"mist/src/errors/message"
	"mist/src/permission"
	"mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *ChannelGRPCService) Create(
	ctx context.Context, req *channel.CreateRequest,
) (*channel.CreateResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}
	cs := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer)
	c, err := cs.Create(qx.CreateChannelParams{Name: req.Name, AppserverID: serverId})

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &channel.CreateResponse{
		Channel: cs.PgTypeToPb(c),
	}, nil

}

func (s *ChannelGRPCService) GetById(
	ctx context.Context, req *channel.GetByIdRequest,
) (*channel.GetByIdResponse, error) {
	var err error

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionRead, permission.SubActionGetById); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	cs := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer)
	id, err := uuid.Parse(req.Id)
	c, err := cs.GetById(id)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
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

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverChannels); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	cs := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer)

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

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	id, _ := uuid.Parse(req.Id)
	if err := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer).Delete(id); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &channel.DeleteResponse{}, nil
}
