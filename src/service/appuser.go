package service

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	ctx  context.Context
	deps *ServiceDeps
}

// Creates a new AppuserService struct.
func NewAppuserService(ctx context.Context, deps *ServiceDeps) *AppuserService {
	return &AppuserService{ctx: ctx, deps: deps}
}

// Convert Appuser db object to Appuser protobuff object.
func (s *AppuserService) PgTypeToPb(a *qx.Appuser) *appuser.Appuser {
	return &appuser.Appuser{
		Id:        a.ID.String(),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

// Creates a new appuser.
func (s *AppuserService) Create(obj qx.CreateAppuserParams) (*qx.Appuser, error) {
	as, err := s.deps.Db.CreateAppuser(s.ctx, obj)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("create appuser: %v", err), slog.LevelError)

	}
	return &as, err

}
