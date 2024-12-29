package rpcs

import (
	"context"

	pb_servers "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateChannel(
	ctx context.Context, req *pb_servers.CreateChannelRequest,
) (*pb_servers.CreateChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.Create(req.GetName(), req.GetAppserverId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_servers.CreateChannelResponse{
		Channel: channelService.PgTypeToPb(channel),
	}, nil
}

func (s *Grpcserver) GetByIdChannel(
	ctx context.Context, req *pb_servers.GetByIdChannelRequest,
) (*pb_servers.GetByIdChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_servers.GetByIdChannelResponse{Channel: channelService.PgTypeToPb(channel)}, nil
}

func (s *Grpcserver) ListChannels(
	ctx context.Context, req *pb_servers.ListChannelsRequest,
) (*pb_servers.ListChannelsResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	// TODO: Handle potential errors that can happen here
	channels, _ := channelService.List(req.GetName(), req.GetAppserverId())

	response := &pb_servers.ListChannelsResponse{}

	// Resize the array to the correct size
	response.Channels = make([]*pb_servers.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, channelService.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *Grpcserver) DeleteChannel(
	ctx context.Context, req *pb_servers.DeleteChannelRequest,
) (*pb_servers.DeleteChannelResponse, error) {

	err := service.NewChannelService(s.DbcPool, ctx).Delete(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_servers.DeleteChannelResponse{}, nil
}
