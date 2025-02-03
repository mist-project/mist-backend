package rpcs

import (
	"context"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/service"
)

func (s *AppuserGRPCService) CreateAppuser(
	ctx context.Context, req *pb_appuser.CreateAppuserRequest,
) (*pb_appuser.CreateAppuserResponse, error) {

	_, err := service.NewAppuserService(s.DbcPool, ctx).Create(req.GetUsername(), req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_appuser.CreateAppuserResponse{}, nil
}
