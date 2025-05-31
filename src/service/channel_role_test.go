package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestChannelRoleService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	id := uuid.New()
	channelId := uuid.New()
	roleId := uuid.New()
	now := time.Now()

	role := &qx.ChannelRole{
		ID:              id,
		ChannelID:       channelId,
		AppserverRoleID: roleId,
		CreatedAt:       pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:       pgtype.Timestamp{Time: now, Valid: true},
	}

	expected := &channel_role.ChannelRole{
		Id:              id.String(),
		ChannelId:       channelId.String(),
		AppserverRoleId: roleId.String(),
		CreatedAt:       timestamppb.New(now),
		UpdatedAt:       timestamppb.New(now),
	}

	svc := service.NewChannelRoleService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier))

	// ACT
	res := svc.PgTypeToPb(role)

	// ASSERT
	assert.Equal(t, expected, res)
}

func TestChannelRoleService_Create(t *testing.T) {
	t.Run("Successful:create_channel_role", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}
		expected := qx.ChannelRole{ID: uuid.New(), ChannelID: obj.ChannelID, AppserverRoleID: obj.AppserverRoleID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannelRole", ctx, obj).Return(expected, nil)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		res, err := svc.Create(obj)

		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannelRole", ctx, obj).Return(qx.ChannelRole{}, fmt.Errorf("creation failed"))

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		_, err := svc.Create(obj)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: creation failed")
	})
}

func TestChannelRoleService_ListChannelRoles(t *testing.T) {
	t.Run("Successful:list_roles", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()
		expected := []qx.ChannelRole{{ID: uuid.New(), ChannelID: channelId, AppserverRoleID: uuid.New()}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(expected, nil)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		roles, err := svc.ListChannelRoles(channelId)

		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(nil, fmt.Errorf("db error"))

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		_, err := svc.ListChannelRoles(channelId)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db error")
	})
}

func TestChannelRoleService_GetById(t *testing.T) {
	t.Run("Successful:returns_channel_role", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()
		expected := qx.ChannelRole{ID: roleId, ChannelID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(expected, nil)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		actual, err := svc.GetById(roleId)

		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
	})

	t.Run("Error:returns_not_found", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf(message.DbNotFound))

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		_, err := svc.GetById(roleId)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", roleId))
	})

	t.Run("Error:returns_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf("boom"))

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		_, err := svc.GetById(roleId)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
	})
}

func TestChannelRoleService_Delete(t *testing.T) {
	t.Run("Successful:delete_role", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannelRole", ctx, id).Return(int64(1), nil)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		err := svc.Delete(id)

		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannelRole", ctx, id).Return(int64(0), nil)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		err := svc.Delete(id)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", id))
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannelRole", ctx, id).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier)
		err := svc.Delete(id)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db crash")
	})
}
