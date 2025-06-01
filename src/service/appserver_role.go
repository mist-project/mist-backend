package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appserver_role"
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

func (s *AppserverRoleService) PgTypeToPb(aRole *qx.AppserverRole) *appserver_role.AppserverRole {
	return &appserver_role.AppserverRole{
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
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &appserverRole, err
}

// Lists all the roles for an appserver.
func (s *AppserverRoleService) ListAppserverRoles(appserverId uuid.UUID) ([]qx.AppserverRole, error) {
	aRoles, err := s.db.ListAppserverRoles(s.ctx, appserverId)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return aRoles, nil
}

// Gets an appserver role by its id.
func (s *AppserverRoleService) GetById(id uuid.UUID) (*qx.AppserverRole, error) {
	role, err := s.db.GetAppserverRoleById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), message.DbNotFound) {
			return nil, faults.NotFoundError(err.Error(), slog.LevelDebug)
		}

		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return &role, nil
}

// Lists all the roles for a user in a server.
func (s *AppserverRoleService) GetAppuserRoles(params qx.GetAppuserRolesParams) ([]qx.GetAppuserRolesRow, error) {
	rows, err := s.db.GetAppuserRoles(s.ctx, params)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	}

	return rows, nil
}

// Deletes a role from a server, only owner of server and delete role
func (s *AppserverRoleService) Delete(id uuid.UUID) error {
	deleted, err := s.db.DeleteAppserverRole(s.ctx, id)

	if err != nil {
		return faults.DatabaseError(fmt.Sprintf("database error: %v", err), slog.LevelError)
	} else if deleted == 0 {
		return faults.NotFoundError(fmt.Sprintf("unable to to find role with id: %v", id), slog.LevelDebug)
	}

	return nil
}
