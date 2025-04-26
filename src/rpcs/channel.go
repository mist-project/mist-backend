package rpcs

import (
	"context"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *ChannelGRPCService) CreateChannel(
	ctx context.Context, req *pb_channel.CreateChannelRequest,
) (*pb_channel.CreateChannelResponse, error) {

	cs := service.NewChannelService(ctx, s.DbConn, s.Db)
	serverId, _ := uuid.Parse(req.AppserverId)
	channel, err := cs.Create(qx.CreateChannelParams{Name: req.Name, AppserverID: serverId})

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.CreateChannelResponse{
		Channel: cs.PgTypeToPb(channel),
	}, nil

}

func (s *ChannelGRPCService) GetByIdChannel(
	ctx context.Context, req *pb_channel.GetByIdChannelRequest,
) (*pb_channel.GetByIdChannelResponse, error) {

	cs := service.NewChannelService(ctx, s.DbConn, s.Db)
	id, err := uuid.Parse(req.Id)
	channel, err := cs.GetById(id)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.GetByIdChannelResponse{Channel: cs.PgTypeToPb(channel)}, nil
}

func (s *ChannelGRPCService) ListChannels(
	ctx context.Context, req *pb_channel.ListChannelsRequest,
) (*pb_channel.ListChannelsResponse, error) {

	cs := service.NewChannelService(ctx, s.DbConn, s.Db)
	var (
		nameFilter   pgtype.Text
		serverFilter pgtype.UUID
	)

	if req.Name != nil {
		nameFilter = pgtype.Text{Valid: true, String: req.Name.Value}
	}

	if req.AppserverId != nil {
		serverId, _ := uuid.Parse(req.AppserverId.Value)
		serverFilter = pgtype.UUID{Valid: true, Bytes: serverId}
	}

	channels, _ := cs.List(qx.ListChannelsParams{Name: nameFilter, AppserverID: serverFilter})
	response := &pb_channel.ListChannelsResponse{}
	response.Channels = make([]*pb_channel.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, cs.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *ChannelGRPCService) DeleteChannel(
	ctx context.Context, req *pb_channel.DeleteChannelRequest,
) (*pb_channel.DeleteChannelResponse, error) {

	id, _ := uuid.Parse(req.Id)
	if err := service.NewChannelService(ctx, s.DbConn, s.Db).Delete(id); err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.DeleteChannelResponse{}, nil
}
