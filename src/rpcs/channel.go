package rpcs

import (
	"context"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
)

func (s *ChannelGRPCService) CreateChannel(
	ctx context.Context, req *pb_channel.CreateChannelRequest,
) (*pb_channel.CreateChannelResponse, error) {

	cs := service.NewChannelService(s.DbConn, ctx)
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

	cs := service.NewChannelService(s.DbConn, ctx)
	id, _ := uuid.Parse(req.Id)
	channel, err := cs.GetById(id)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.GetByIdChannelResponse{Channel: cs.PgTypeToPb(channel)}, nil
}

func (s *ChannelGRPCService) ListChannels(
	ctx context.Context, req *pb_channel.ListChannelsRequest,
) (*pb_channel.ListChannelsResponse, error) {

	cs := service.NewChannelService(s.DbConn, ctx)
	// TODO: Handle potential errors that can happen here
	channels, _ := cs.List(req.GetName(), req.GetAppserverId())

	response := &pb_channel.ListChannelsResponse{}

	// Resize the array to the correct size
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
	if err := service.NewChannelService(s.DbConn, ctx).Delete(id); err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.DeleteChannelResponse{}, nil
}
