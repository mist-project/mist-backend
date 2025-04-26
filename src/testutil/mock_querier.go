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

func (m *MockQuerier) WithTx(tx pgx.Tx) db.Querier {
	args := m.Called(tx)
	return args.Get(0).(db.Querier)
}

func (m *MockQuerier) CreateAppserver(ctx context.Context, arg qx.CreateAppserverParams) (qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.Appserver), args.Error(1)
}

func (m *MockQuerier) CreateAppserverRole(ctx context.Context, arg qx.CreateAppserverRoleParams) (qx.AppserverRole, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.AppserverRole), args.Error(1)
}

func (m *MockQuerier) CreateAppserverRoleSub(ctx context.Context, arg qx.CreateAppserverRoleSubParams) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.AppserverRoleSub), args.Error(1)
}

func (m *MockQuerier) CreateAppserverSub(ctx context.Context, arg qx.CreateAppserverSubParams) (qx.AppserverSub, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.AppserverSub), args.Error(1)
}

func (m *MockQuerier) CreateAppuser(ctx context.Context, arg qx.CreateAppuserParams) (qx.Appuser, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.Appuser), args.Error(1)
}

func (m *MockQuerier) CreateChannel(ctx context.Context, arg qx.CreateChannelParams) (qx.Channel, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(qx.Channel), args.Error(1)
}

func (m *MockQuerier) DeleteAppserver(ctx context.Context, arg qx.DeleteAppserverParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) DeleteAppserverRole(ctx context.Context, arg qx.DeleteAppserverRoleParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) DeleteAppserverRoleSub(ctx context.Context, arg qx.DeleteAppserverRoleSubParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) DeleteAppserverSub(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) DeleteChannel(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetAllUsersAppserverSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.GetAllUsersAppserverSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return args.Get(0).([]qx.GetAllUsersAppserverSubsRow), args.Error(1)
}

func (m *MockQuerier) GetAppserverAllUserRoleSubs(ctx context.Context, appserverID uuid.UUID) ([]qx.GetAppserverAllUserRoleSubsRow, error) {
	args := m.Called(ctx, appserverID)
	return args.Get(0).([]qx.GetAppserverAllUserRoleSubsRow), args.Error(1)
}

func (m *MockQuerier) GetAppserverById(ctx context.Context, id uuid.UUID) (qx.Appserver, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.Appserver), args.Error(1)
}

func (m *MockQuerier) GetAppserverRoleById(ctx context.Context, id uuid.UUID) (qx.AppserverRole, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.AppserverRole), args.Error(1)
}

func (m *MockQuerier) GetAppserverRoleSubById(ctx context.Context, id uuid.UUID) (qx.AppserverRoleSub, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.AppserverRoleSub), args.Error(1)
}

func (m *MockQuerier) GetAppserverRoles(ctx context.Context, appserverID uuid.UUID) ([]qx.AppserverRole, error) {
	args := m.Called(ctx, appserverID)
	return args.Get(0).([]qx.AppserverRole), args.Error(1)
}

func (m *MockQuerier) GetAppserverSubById(ctx context.Context, id uuid.UUID) (qx.AppserverSub, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.AppserverSub), args.Error(1)
}

func (m *MockQuerier) GetAppuserById(ctx context.Context, id uuid.UUID) (qx.Appuser, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.Appuser), args.Error(1)
}

func (m *MockQuerier) GetAppuserRoleSubs(ctx context.Context, arg qx.GetAppuserRoleSubsParams) ([]qx.GetAppuserRoleSubsRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]qx.GetAppuserRoleSubsRow), args.Error(1)
}

func (m *MockQuerier) GetChannelById(ctx context.Context, id uuid.UUID) (qx.Channel, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(qx.Channel), args.Error(1)
}

func (m *MockQuerier) GetUserAppserverSubs(ctx context.Context, appuserID uuid.UUID) ([]qx.GetUserAppserverSubsRow, error) {
	args := m.Called(ctx, appuserID)
	return args.Get(0).([]qx.GetUserAppserverSubsRow), args.Error(1)
}

func (m *MockQuerier) ListServerChannels(ctx context.Context, arg qx.ListServerChannelsParams) ([]qx.Channel, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]qx.Channel), args.Error(1)
}

func (m *MockQuerier) ListUserAppservers(ctx context.Context, arg qx.ListUserAppserversParams) ([]qx.Appserver, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]qx.Appserver), args.Error(1)
}
