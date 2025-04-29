package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	pb_appserverpermission "mist/src/protos/v1/appserver_permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverPermissionService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverPermissionService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverPermissionService {
	return &AppserverPermissionService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverPermissionService) PgTypeToPb(p *qx.AppserverPermission) *pb_appserverpermission.AppserverPermission {
	return &pb_appserverpermission.AppserverPermission{
		Id:          p.ID.String(),
		AppserverId: p.AppserverID.String(),
		AppuserId:   p.AppuserID.String(),
		CreatedAt:   timestamppb.New(p.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(p.UpdatedAt.Time),
	}
}

// Creates an appserver permission.
func (s *AppserverPermissionService) Create(obj qx.CreateAppserverPermissionParams) (*qx.AppserverPermission, error) {
	appserverRole, err := s.db.CreateAppserverPermission(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &appserverRole, err
}

// Lists all users with appserver permission for an appserver.
func (s *AppserverPermissionService) ListAppserverPermissions(appserverId uuid.UUID) ([]qx.AppserverPermission, error) {
	aRoles, err := s.db.ListAppserverPermissions(s.ctx, appserverId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return aRoles, nil
}

// Gets an appserver permission by its id.
func (s *AppserverPermissionService) GetById(id uuid.UUID) (*qx.AppserverPermission, error) {
	role, err := s.db.GetAppserverPermissionById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, message.NotFoundError(message.NotFound)
		}

		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &role, nil
}

// Deletes a permission from a server, only owner of server and delete permission
func (s *AppserverPermissionService) Delete(id uuid.UUID) error {
	deleted, err := s.db.DeleteAppserverPermission(s.ctx, id)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}
	return nil
}
