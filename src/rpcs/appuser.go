package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/faults"
	"mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppuserGRPCService) Create(
	ctx context.Context, req *appuser.CreateRequest,
) (*appuser.CreateResponse, error) {

	userId, _ := uuid.Parse(req.Id)
	_, err := service.NewAppuserService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).Create(
		qx.CreateAppuserParams{
			ID:       userId,
			Username: req.Username,
		},
	)

	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	return &appuser.CreateResponse{}, nil
}
