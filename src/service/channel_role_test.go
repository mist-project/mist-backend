package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/channel_role"
	"mist/src/protos/v1/event"
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

	mockProducer := new(testutil.MockProducer)
	svc := service.NewChannelRoleService(context.Background(), testutil.TestDbConn, new(testutil.MockQuerier), mockProducer)

	// ACT
	res := svc.PgTypeToPb(role)

	// ASSERT
	assert.Equal(t, expected, res)
}

func TestChannelRoleService_Create(t *testing.T) {
	t.Run("Successful:create_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}
		expected := qx.ChannelRole{ID: uuid.New(), ChannelID: obj.ChannelID, AppserverRoleID: obj.AppserverRoleID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannelRole", ctx, obj).Return(expected, nil)
		mockProducer := new(testutil.MockProducer)

		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannelRole", ctx, obj).Return(qx.ChannelRole{}, fmt.Errorf("creation failed"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: creation failed")
	})
}

func TestChannelRoleService_ListChannelRoles(t *testing.T) {
	t.Run("Successful:list_roles", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()
		expected := []qx.ChannelRole{{ID: uuid.New(), ChannelID: channelId, AppserverRoleID: uuid.New()}}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(expected, nil)
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		roles, err := svc.ListChannelRoles(channelId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(nil, fmt.Errorf("db error"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.ListChannelRoles(channelId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db error")
	})
}

func TestChannelRoleService_GetById(t *testing.T) {
	t.Run("Successful:returns_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()
		expected := qx.ChannelRole{ID: roleId, ChannelID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(expected, nil)
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		actual, err := svc.GetById(roleId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
	})

	t.Run("Error:returns_not_found", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf(message.DbNotFound))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.GetById(roleId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", roleId))
	})

	t.Run("Error:returns_db_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf("boom"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.GetById(roleId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
	})
}

func TestChannelRoleService_Delete(t *testing.T) {
	t.Run("Successful:delete_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(int64(1), nil)
		mockQuerier.On("GetAppusersWithOnlySpecifiedRole", ctx, channelRole.AppserverRoleID).Return([]qx.Appuser{
			{ID: uuid.New(), Username: "testuser"},
		}, nil)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_CHANNEL, mock.Anything).Return(nil)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(int64(0), nil)
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", channelRole.ID))
	})

	t.Run("Error:db_failure_on_get_channel_role_by_id", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(nil, fmt.Errorf("boom"))
		// mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, nil)
		// mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(nil, fmt.Errorf("db crash"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(nil, fmt.Errorf("db crash"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db crash")
	})

	t.Run("Error:db_failure_on_get_channel_by_id", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(nil, fmt.Errorf("boom"))
		mockProducer := new(testutil.MockProducer)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
	})

	t.Run("Successful:delete_role_is_successful_even_on_error_on_get_appuser_with_specified_role", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(int64(1), nil)
		mockQuerier.On("GetAppusersWithOnlySpecifiedRole", ctx, channelRole.AppserverRoleID).Return(nil, fmt.Errorf("boom"))
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_CHANNEL, mock.Anything).Return(nil)
		svc := service.NewChannelRoleService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.NoError(t, err)
	})

}
