package rpcs

import (
	"context"

	"github.com/google/uuid"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppuserGRPCService) CreateAppuser(
	ctx context.Context, req *pb_appuser.CreateAppuserRequest,
) (*pb_appuser.CreateAppuserResponse, error) {

	userId, _ := uuid.Parse(req.Id)
	_, err := service.NewAppuserService(ctx, s.DbConn, s.Db).Create(
		qx.CreateAppuserParams{
			ID:       userId,
			Username: req.Username,
		},
	)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_appuser.CreateAppuserResponse{}, nil
}
