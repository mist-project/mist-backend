package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	dbConn qx.DBTX
	ctx    context.Context
}

func NewAppuserService(dbConn qx.DBTX, ctx context.Context) *AppuserService {
	return &AppuserService{dbConn: dbConn, ctx: ctx}
}

func (s *AppuserService) PgTypeToPb(a *qx.Appuser) *pb_appuser.Appuser {
	return &pb_appuser.Appuser{
		Id:        a.ID.String(),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

func (s *AppuserService) Create(obj qx.CreateAppuserParams) (*qx.Appuser, error) {
	as, err := qx.New(s.dbConn).CreateAppuser(s.ctx, obj)
	return &as, err
}
