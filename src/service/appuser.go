package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	ctx context.Context
	db  qx.Querier
}

func NewAppuserService(db qx.Querier, ctx context.Context) *AppuserService {
	return &AppuserService{db: db, ctx: ctx}
}

func (s *AppuserService) PgTypeToPb(a *qx.Appuser) *pb_appuser.Appuser {
	return &pb_appuser.Appuser{
		Id:        a.ID.String(),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppuserService) Create(obj qx.CreateAppuserParams) (*qx.Appuser, error) {
	as, err := s.db.CreateAppuser(s.ctx, obj)
	return &as, err
}
