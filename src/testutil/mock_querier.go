package testutil

import (
	"context"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) Begin(ctx context.Context) (db.Querier, error) {
	args := m.Called(ctx)

	err := args.Error(1)
	if err != nil {
		return nil, err
	}

	return args.Get(0).(db.Querier), nil
}

func (m *MockQuerier) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Rollback(ctx context.Context) error
func (m *MockQuerier) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *MockQuerier) CreateAppserver(ctx context.Context, arg qx.CreateAppserverParams) (qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.Appserver](args, 1)
}

func (m *MockQuerier) CreateAppserverRole(ctx context.Context, arg qx.CreateAppserverRoleParams) (qx.AppserverRole, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.AppserverRole](args, 1)
}

func (m *MockQuerier) CreateAppserverRoleSub(ctx context.Context, arg qx.CreateAppserverRoleSubParams) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.AppserverRoleSub](args, 1)
}

func (m *MockQuerier) CreateAppserverSub(ctx context.Context, arg qx.CreateAppserverSubParams) (qx.AppserverSub, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.AppserverSub](args, 1)
}

func (m *MockQuerier) CreateAppuser(ctx context.Context, arg qx.CreateAppuserParams) (qx.Appuser, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.Appuser](args, 1)
}

func (m *MockQuerier) CreateChannel(ctx context.Context, arg qx.CreateChannelParams) (qx.Channel, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.Channel](args, 1)
}

func (m *MockQuerier) CreateChannelRole(ctx context.Context, arg qx.CreateChannelRoleParams) (qx.ChannelRole, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[qx.ChannelRole](args, 1)
}

func (m *MockQuerier) DeleteAppserver(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverPermission(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverRole(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverRoleSub(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteAppserverSub(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannel(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannelPermission(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) DeleteChannelRole(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[int64](args, 1)
}

func (m *MockQuerier) FilterAppserverSub(ctx context.Context, arg qx.FilterAppserverSubParams) ([]qx.FilterAppserverSubRow, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.FilterAppserverSubRow](args, 1)
}

func (m *MockQuerier) FilterChannelRole(ctx context.Context, arg qx.FilterChannelRoleParams) ([]qx.FilterChannelRoleRow, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.FilterChannelRoleRow](args, 1)
}

func (m *MockQuerier) FilterChannel(ctx context.Context, arg qx.FilterChannelParams) ([]qx.Channel, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.Channel](args, 1)
}

func (m *MockQuerier) FilterAppserverRoleSub(ctx context.Context, arg qx.FilterAppserverRoleSubParams) ([]qx.FilterAppserverRoleSubRow, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.FilterAppserverRoleSubRow](args, 1)
}

func (m *MockQuerier) ListAppserverUserSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.ListAppserverUserSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return ReturnIfError[[]qx.ListAppserverUserSubsRow](args, 1)
}

func (m *MockQuerier) ListServerRoleSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.ListServerRoleSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return ReturnIfError[[]qx.ListServerRoleSubsRow](args, 1)
}

func (m *MockQuerier) GetAppserverById(ctx context.Context, id uuid.UUID) (qx.Appserver, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.Appserver](args, 1)
}

func (m *MockQuerier) GetAppserverRoleById(ctx context.Context, id uuid.UUID) (qx.AppserverRole, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.AppserverRole](args, 1)
}

func (m *MockQuerier) GetAppserverRoleSubById(ctx context.Context, id uuid.UUID) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.AppserverRoleSub](args, 1)
}

func (m *MockQuerier) GetAppusersWithOnlySpecifiedRole(ctx context.Context, appserverRoleID uuid.UUID) ([]qx.Appuser, error) {
	args := m.Called(ctx, appserverRoleID)
	return ReturnIfError[[]qx.Appuser](args, 1)
}

func (m *MockQuerier) GetChannelRoleById(ctx context.Context, id uuid.UUID) (qx.ChannelRole, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.ChannelRole](args, 1)
}

func (q *MockQuerier) GetChannelsForUsers(ctx context.Context, arg qx.GetChannelsForUsersParams) ([]qx.GetChannelsForUsersRow, error) {
	args := q.Called(ctx, arg)
	return ReturnIfError[[]qx.GetChannelsForUsersRow](args, 1)
}

func (q *MockQuerier) GetChannelsIdIn(ctx context.Context, dollar_1 []uuid.UUID) ([]qx.Channel, error) {
	args := q.Called(ctx, dollar_1)
	return ReturnIfError[[]qx.Channel](args, 1)
}

func (m *MockQuerier) ListAppserverRoles(ctx context.Context, appserverID uuid.UUID) ([]qx.AppserverRole, error) {
	args := m.Called(ctx, appserverID)
	return ReturnIfError[[]qx.AppserverRole](args, 1)
}

func (m *MockQuerier) GetAppserverSubById(ctx context.Context, id uuid.UUID) (qx.AppserverSub, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.AppserverSub](args, 1)
}

func (m *MockQuerier) GetAppuserById(ctx context.Context, id uuid.UUID) (qx.Appuser, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.Appuser](args, 1)
}

func (m *MockQuerier) GetAppuserRoles(ctx context.Context, arg qx.GetAppuserRolesParams) ([]qx.GetAppuserRolesRow, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.GetAppuserRolesRow](args, 1)
}

func (m *MockQuerier) ListChannelRoles(ctx context.Context, id uuid.UUID) ([]qx.ChannelRole, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[[]qx.ChannelRole](args, 1)
}

func (m *MockQuerier) GetChannelById(ctx context.Context, id uuid.UUID) (qx.Channel, error) {
	args := m.Called(ctx, id)
	return ReturnIfError[qx.Channel](args, 1)
}

func (m *MockQuerier) ListUserServerSubs(ctx context.Context, appuserID uuid.UUID) ([]qx.ListUserServerSubsRow, error) {
	args := m.Called(ctx, appuserID)
	return ReturnIfError[[]qx.ListUserServerSubsRow](args, 1)
}

func (m *MockQuerier) ListServerChannels(ctx context.Context, arg qx.ListServerChannelsParams) ([]qx.Channel, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.Channel](args, 1)
}

func (m *MockQuerier) ListAppservers(ctx context.Context, arg qx.ListAppserversParams) ([]qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return ReturnIfError[[]qx.Appserver](args, 1)
}
