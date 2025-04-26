package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppuserService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppuserService {
	return &AppuserService{ctx: ctx, dbConn: dbConn, db: db}
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
