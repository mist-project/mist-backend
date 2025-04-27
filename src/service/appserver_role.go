package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverRoleService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewAppserverRoleService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverRoleService {
	return &AppserverRoleService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *AppserverRoleService) PgTypeToPb(aRole *qx.AppserverRole) *pb_appserverrole.AppserverRole {
	return &pb_appserverrole.AppserverRole{
		Id:          aRole.ID.String(),
		AppserverId: aRole.AppserverID.String(),
		Name:        aRole.Name,
		CreatedAt:   timestamppb.New(aRole.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(aRole.UpdatedAt.Time),
	}
}

// Creates an appserver role.
func (s *AppserverRoleService) Create(obj qx.CreateAppserverRoleParams) (*qx.AppserverRole, error) {
	appserverRole, err := s.db.CreateAppserverRole(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &appserverRole, err
}

// Lists all the roles for an appserver.
func (s *AppserverRoleService) ListAppserverRoles(appserverId uuid.UUID) ([]qx.AppserverRole, error) {
	aRoles, err := s.db.ListAppserverRoles(s.ctx, appserverId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return aRoles, nil
}

// Gets an appserver detail by its id.
func (s *AppserverRoleService) GetById(id uuid.UUID) (*qx.AppserverRole, error) {
	role, err := s.db.GetAppserverRoleById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, message.NotFoundError(message.NotFound)
		}

		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &role, nil
}

// Deletes a role from a server, only owner of server and delete role
func (s *AppserverRoleService) Delete(obj qx.DeleteAppserverRoleParams) error {
	deleted, err := s.db.DeleteAppserverRole(s.ctx, obj)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}
	return nil
}
