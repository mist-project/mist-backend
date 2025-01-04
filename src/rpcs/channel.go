package rpcs

import (
	"context"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/service"
)

func (s *ChannelGRPCService) CreateChannel(
	ctx context.Context, req *pb_channel.CreateChannelRequest,
) (*pb_channel.CreateChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.Create(req.GetName(), req.GetAppserverId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.CreateChannelResponse{
		Channel: channelService.PgTypeToPb(channel),
	}, nil
}

func (s *ChannelGRPCService) GetByIdChannel(
	ctx context.Context, req *pb_channel.GetByIdChannelRequest,
) (*pb_channel.GetByIdChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.GetByIdChannelResponse{Channel: channelService.PgTypeToPb(channel)}, nil
}

func (s *ChannelGRPCService) ListChannels(
	ctx context.Context, req *pb_channel.ListChannelsRequest,
) (*pb_channel.ListChannelsResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	// TODO: Handle potential errors that can happen here
	channels, _ := channelService.List(req.GetName(), req.GetAppserverId())

	response := &pb_channel.ListChannelsResponse{}

	// Resize the array to the correct size
	response.Channels = make([]*pb_channel.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, channelService.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *ChannelGRPCService) DeleteChannel(
	ctx context.Context, req *pb_channel.DeleteChannelRequest,
) (*pb_channel.DeleteChannelResponse, error) {

	err := service.NewChannelService(s.DbcPool, ctx).Delete(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_channel.DeleteChannelResponse{}, nil
}
