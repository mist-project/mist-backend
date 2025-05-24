package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/errors/message"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type ChannelRoleService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

func NewChannelRoleService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *ChannelRoleService {
	return &ChannelRoleService{ctx: ctx, dbConn: dbConn, db: db}
}

func (s *ChannelRoleService) PgTypeToPb(cRole *qx.ChannelRole) *channel_role.ChannelRole {
	return &channel_role.ChannelRole{
		Id:              cRole.ID.String(),
		ChannelId:       cRole.ChannelID.String(),
		AppserverRoleId: cRole.AppserverRoleID.String(),
		CreatedAt:       timestamppb.New(cRole.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(cRole.UpdatedAt.Time),
	}
}

// Creates an appserver role.
func (s *ChannelRoleService) Create(obj qx.CreateChannelRoleParams) (*qx.ChannelRole, error) {
	appserverRole, err := s.db.CreateChannelRole(s.ctx, obj)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return &appserverRole, err
}

// Lists all the roles for an appserver.
func (s *ChannelRoleService) ListChannelRoles(channelId uuid.UUID) ([]qx.ChannelRole, error) {
	cRoles, err := s.db.ListChannelRoles(s.ctx, channelId)

	if err != nil {
		return nil, message.DatabaseError(fmt.Sprintf("database error: %v", err))
	}

	return cRoles, nil
}

// Gets an appserver role by its id.
func (s *ChannelRoleService) GetById(id uuid.UUID) (*qx.ChannelRole, error) {
	role, err := s.db.GetChannelRoleById(s.ctx, id)

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
func (s *ChannelRoleService) Delete(id uuid.UUID) error {
	deleted, err := s.db.DeleteChannelRole(s.ctx, id)

	if err != nil {
		return message.DatabaseError(fmt.Sprintf("database error: %v", err))
	} else if deleted == 0 {
		return message.NotFoundError(message.NotFound)
	}
	return nil
}
