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
	"mist/src/producer"
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

	svc := service.NewChannelRoleService(
		context.Background(),
		&service.ServiceDeps{
			Db:        new(testutil.MockQuerier),
			MProducer: producer.NewMProducer(new(testutil.MockRedis)),
		})

	// ACT
	res := svc.PgTypeToPb(role)

	// ASSERT
	assert.Equal(t, expected, res)
}

func TestChannelRoleService_Create(t *testing.T) {
	t.Run("Success:create_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}
		expected := qx.ChannelRole{ID: uuid.New(), ChannelID: obj.ChannelID, AppserverRoleID: obj.AppserverRoleID}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateChannelRole", ctx, obj).Return(expected, nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, obj.AppserverID).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		res, err := svc.Create(obj)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, res.ID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_create_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		obj := qx.CreateChannelRoleParams{ChannelID: uuid.New(), AppserverRoleID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateChannelRole", ctx, obj).Return(nil, fmt.Errorf("creation failed"))
		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.Create(obj)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: creation failed")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelRoleService_ListChannelRoles(t *testing.T) {
	t.Run("Success:list_roles", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channelId := uuid.New()
		expected := []qx.ChannelRole{{ID: uuid.New(), ChannelID: channelId, AppserverRoleID: uuid.New()}}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(expected, nil)

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		roles, err := svc.ListChannelRoles(channelId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, roles)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channelId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListChannelRoles", ctx, channelId).Return(nil, fmt.Errorf("db error"))

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.ListChannelRoles(channelId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db error")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelRoleService_GetById(t *testing.T) {
	t.Run("Success:returns_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleId := uuid.New()
		expected := qx.ChannelRole{ID: roleId, ChannelID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(expected, nil)

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		actual, err := svc.GetById(roleId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:returns_not_found", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf(message.DbNotFound))

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(roleId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", roleId))
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:returns_db_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		roleId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, roleId).Return(qx.ChannelRole{}, fmt.Errorf("boom"))

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(roleId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelRoleService_Delete(t *testing.T) {
	t.Run("Success:delete_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New()}
		channelRole := qx.ChannelRole{
			ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID, AppserverID: uuid.New(),
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, channelRole.AppserverID).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.NoError(t, err)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRole := qx.ChannelRole{ID: uuid.New(), AppserverRoleID: uuid.New(), ChannelID: channel.ID}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(int64(0), nil)

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, fmt.Sprintf("unable to find channel role with id: %v", channelRole.ID))
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:db_failure_on_get_channel_role_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channelRole := qx.ChannelRole{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(nil, fmt.Errorf("boom"))

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: boom")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:db_failure_on_delete", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channelRole := qx.ChannelRole{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelRoleById", ctx, channelRole.ID).Return(channelRole, nil)
		mockQuerier.On("DeleteChannelRole", ctx, channelRole.ID).Return(nil, fmt.Errorf("db crash"))

		svc := service.NewChannelRoleService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(channelRole.ID)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db crash")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}
