package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppuserGRPCService) Create(
	ctx context.Context, req *pb_appuser.CreateRequest,
) (*pb_appuser.CreateResponse, error) {

	userId, _ := uuid.Parse(req.Id)
	_, err := service.NewAppuserService(ctx, s.DbConn, s.Db).Create(
		qx.CreateAppuserParams{
			ID:       userId,
			Username: req.Username,
		},
	)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &pb_appuser.CreateResponse{}, nil
}
