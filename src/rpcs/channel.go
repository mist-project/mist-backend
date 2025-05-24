package rpcs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"mist/src/errors/message"
	"mist/src/permission"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *ChannelGRPCService) Create(
	ctx context.Context, req *pb_channel.CreateRequest,
) (*pb_channel.CreateResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}
	cs := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer)
	channel, err := cs.Create(qx.CreateChannelParams{Name: req.Name, AppserverID: serverId})

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &pb_channel.CreateResponse{
		Channel: cs.PgTypeToPb(channel),
	}, nil

}

func (s *ChannelGRPCService) GetById(
	ctx context.Context, req *pb_channel.GetByIdRequest,
) (*pb_channel.GetByIdResponse, error) {
	var err error

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionRead, permission.SubActionGetById); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	cs := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer)
	id, err := uuid.Parse(req.Id)
	channel, err := cs.GetById(id)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &pb_channel.GetByIdResponse{Channel: cs.PgTypeToPb(channel)}, nil
}

func (s *ChannelGRPCService) ListServerChannels(
	ctx context.Context, req *pb_channel.ListServerChannelsRequest,
) (*pb_channel.ListServerChannelsResponse, error) {

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
	response := &pb_channel.ListServerChannelsResponse{}
	response.Channels = make([]*pb_channel.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, cs.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *ChannelGRPCService) Delete(
	ctx context.Context, req *pb_channel.DeleteRequest,
) (*pb_channel.DeleteResponse, error) {

	var err error

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	id, _ := uuid.Parse(req.Id)
	if err := service.NewChannelService(ctx, s.DbConn, s.Db, s.Producer).Delete(id); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &pb_channel.DeleteResponse{}, nil
}
