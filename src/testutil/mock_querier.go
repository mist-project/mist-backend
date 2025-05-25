package testutil

import (
	"context"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func returnIfError[T any](args mock.Arguments, index int) (T, error) {
	err := args.Error(index)
	var zero T
	if err != nil {
		return zero, err
	}
	return args.Get(0).(T), nil
}

func (m *MockQuerier) WithTx(tx pgx.Tx) db.Querier {
	args := m.Called(tx)
	return args.Get(0).(db.Querier)
}

func (m *MockQuerier) CreateAppserver(ctx context.Context, arg qx.CreateAppserverParams) (qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.Appserver](args, 1)
}

func (m *MockQuerier) CreateAppserverPermission(ctx context.Context, arg qx.CreateAppserverPermissionParams) (qx.AppserverPermission, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.AppserverPermission](args, 1)
}

func (m *MockQuerier) CreateAppserverRole(ctx context.Context, arg qx.CreateAppserverRoleParams) (qx.AppserverRole, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.AppserverRole](args, 1)
}

func (m *MockQuerier) CreateAppserverRoleSub(ctx context.Context, arg qx.CreateAppserverRoleSubParams) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.AppserverRoleSub](args, 1)
}

func (m *MockQuerier) CreateAppserverSub(ctx context.Context, arg qx.CreateAppserverSubParams) (qx.AppserverSub, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.AppserverSub](args, 1)
}

func (m *MockQuerier) CreateAppuser(ctx context.Context, arg qx.CreateAppuserParams) (qx.Appuser, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.Appuser](args, 1)
}

func (m *MockQuerier) CreateChannel(ctx context.Context, arg qx.CreateChannelParams) (qx.Channel, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.Channel](args, 1)
}

func (m *MockQuerier) CreateChannelPermission(ctx context.Context, arg qx.CreateChannelPermissionParams) (qx.ChannelPermission, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.ChannelPermission](args, 1)
}

func (m *MockQuerier) CreateChannelRole(ctx context.Context, arg qx.CreateChannelRoleParams) (qx.ChannelRole, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.ChannelRole](args, 1)
}

func (m *MockQuerier) DeleteAppserver(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverPermission(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverRole(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverRoleSub(ctx context.Context, arg qx.DeleteAppserverRoleSubParams) (int64, error) {
	args := m.Called(ctx, arg)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverSub(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannel(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannelPermission(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannelRole(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return returnIfError[int64](args, 1)
}

func (m *MockQuerier) FilterAppserverSub(ctx context.Context, arg qx.FilterAppserverSubParams) ([]qx.FilterAppserverSubRow, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.FilterAppserverSubRow](args, 1)
}

func (m *MockQuerier) FilterChannelRole(ctx context.Context, arg qx.FilterChannelRoleParams) ([]qx.FilterChannelRoleRow, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.FilterChannelRoleRow](args, 1)
}

func (m *MockQuerier) ListAppserverUserSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.ListAppserverUserSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return returnIfError[[]qx.ListAppserverUserSubsRow](args, 1)
}

func (m *MockQuerier) ListServerRoleSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.ListServerRoleSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return returnIfError[[]qx.ListServerRoleSubsRow](args, 1)
}

func (m *MockQuerier) GetAppserverById(ctx context.Context, id uuid.UUID) (qx.Appserver, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.Appserver](args, 1)
}

func (m *MockQuerier) GetAppserverPermissionById(ctx context.Context, id uuid.UUID) (qx.AppserverPermission, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.AppserverPermission](args, 1)
}

func (m *MockQuerier) GetAppserverPermissionForUser(ctx context.Context, arg qx.GetAppserverPermissionForUserParams) (qx.AppserverPermission, error) {
	args := m.Called(ctx, arg)
	return returnIfError[qx.AppserverPermission](args, 1)
}

func (m *MockQuerier) GetAppserverRoleById(ctx context.Context, id uuid.UUID) (qx.AppserverRole, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.AppserverRole](args, 1)
}

func (m *MockQuerier) GetAppserverRoleSubById(ctx context.Context, id uuid.UUID) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.AppserverRoleSub](args, 1)
}

func (m *MockQuerier) GetChannelRoleById(ctx context.Context, id uuid.UUID) (qx.ChannelRole, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.ChannelRole](args, 1)
}

func (q *MockQuerier) GetChannelsForUser(ctx context.Context, arg qx.GetChannelsForUserParams) ([]qx.Channel, error) {
	args := q.Called(ctx, arg)
	return returnIfError[[]qx.Channel](args, 1)
}

func (m *MockQuerier) GetChannelUsersByRoles(ctx context.Context, arg []uuid.UUID) ([]qx.Appuser, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.Appuser](args, 1)
}

func (m *MockQuerier) ListAppserverPermissions(ctx context.Context, appserverID uuid.UUID) ([]qx.AppserverPermission, error) {
	args := m.Called(ctx, appserverID)
	return returnIfError[[]qx.AppserverPermission](args, 1)
}

func (m *MockQuerier) ListAppserverRoles(ctx context.Context, appserverID uuid.UUID) ([]qx.AppserverRole, error) {
	args := m.Called(ctx, appserverID)
	return returnIfError[[]qx.AppserverRole](args, 1)
}

func (m *MockQuerier) GetAppserverSubById(ctx context.Context, id uuid.UUID) (qx.AppserverSub, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.AppserverSub](args, 1)
}

func (m *MockQuerier) GetAppuserById(ctx context.Context, id uuid.UUID) (qx.Appuser, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.Appuser](args, 1)
}

func (m *MockQuerier) GetChannelPermissionById(ctx context.Context, id uuid.UUID) (qx.ChannelPermission, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.ChannelPermission](args, 1)
}

func (m *MockQuerier) GetAppuserRoles(ctx context.Context, arg qx.GetAppuserRolesParams) ([]qx.GetAppuserRolesRow, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.GetAppuserRolesRow](args, 1)
}

func (m *MockQuerier) ListChannelPermissions(ctx context.Context, id uuid.UUID) ([]qx.ChannelPermission, error) {
	args := m.Called(ctx, id)
	return returnIfError[[]qx.ChannelPermission](args, 1)
}

func (m *MockQuerier) ListChannelRoles(ctx context.Context, id uuid.UUID) ([]qx.ChannelRole, error) {
	args := m.Called(ctx, id)
	return returnIfError[[]qx.ChannelRole](args, 1)
}

func (m *MockQuerier) GetChannelById(ctx context.Context, id uuid.UUID) (qx.Channel, error) {
	args := m.Called(ctx, id)
	return returnIfError[qx.Channel](args, 1)
}

func (m *MockQuerier) ListUserServerSubs(ctx context.Context, appuserID uuid.UUID) ([]qx.ListUserServerSubsRow, error) {
	args := m.Called(ctx, appuserID)
	return returnIfError[[]qx.ListUserServerSubsRow](args, 1)
}

func (m *MockQuerier) ListServerChannels(ctx context.Context, arg qx.ListServerChannelsParams) ([]qx.Channel, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.Channel](args, 1)
}

func (m *MockQuerier) ListAppservers(ctx context.Context, arg qx.ListAppserversParams) ([]qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return returnIfError[[]qx.Appserver](args, 1)
}
