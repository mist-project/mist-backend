package rpcs

import (
	"context"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateChannel(
	ctx context.Context, req *pb_mistbe.CreateChannelRequest,
) (*pb_mistbe.CreateChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.Create(req.GetName(), req.GetAppserverId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.CreateChannelResponse{
		Channel: channelService.PgTypeToPb(channel),
	}, nil
}

func (s *Grpcserver) GetByIdChannel(
	ctx context.Context, req *pb_mistbe.GetByIdChannelRequest,
) (*pb_mistbe.GetByIdChannelResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	channel, err := channelService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.GetByIdChannelResponse{Channel: channelService.PgTypeToPb(channel)}, nil
}

func (s *Grpcserver) ListChannels(
	ctx context.Context, req *pb_mistbe.ListChannelsRequest,
) (*pb_mistbe.ListChannelsResponse, error) {
	channelService := service.NewChannelService(s.DbcPool, ctx)
	// TODO: Handle potential errors that can happen here
	channels, _ := channelService.List(req.GetName(), req.GetAppserverId())

	response := &pb_mistbe.ListChannelsResponse{}

	// Resize the array to the correct size
	response.Channels = make([]*pb_mistbe.Channel, 0, len(channels))

	for _, channel := range channels {
		response.Channels = append(response.Channels, channelService.PgTypeToPb(&channel))
	}

	return response, nil
}

func (s *Grpcserver) DeleteChannel(
	ctx context.Context, req *pb_mistbe.DeleteChannelRequest,
) (*pb_mistbe.DeleteChannelResponse, error) {

	err := service.NewChannelService(s.DbcPool, ctx).Delete(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.DeleteChannelResponse{}, nil
}
