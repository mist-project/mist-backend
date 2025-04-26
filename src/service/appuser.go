package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppuserService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

// Creates a new AppuserService struct.
func NewAppuserService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppuserService {
	return &AppuserService{ctx: ctx, dbConn: dbConn, db: db}
}

// Convert Appuser db object to Appuser protobuff object.
func (s *AppuserService) PgTypeToPb(a *qx.Appuser) *pb_appuser.Appuser {
	return &pb_appuser.Appuser{
		Id:        a.ID.String(),
		Username:  a.Username,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

// Creates a new appuser.
func (s *AppuserService) Create(obj qx.CreateAppuserParams) (*qx.Appuser, error) {
	as, err := s.db.CreateAppuser(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("create appuser: %v", err))

	}
	return &as, err

}
