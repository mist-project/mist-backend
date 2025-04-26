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

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestChannelService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	svc := service.NewChannelService(ctx, testutil.TestDbConn, new(testutil.MockQuerier))

	id := uuid.New()
	serverId := uuid.New()
	now := time.Now()

	appuser := &qx.Channel{
		ID:          id,
		Name:        "test channel",
		AppserverID: serverId,
		CreatedAt: pgtype.Timestamp{
			Time:  now,
			Valid: true,
		},
	}

	expected := &pb_channel.Channel{
		Id:          id.String(),
		Name:        "test channel",
		CreatedAt:   timestamppb.New(now),
		AppserverId: serverId.String(),
	}

	// ACT
	result := svc.PgTypeToPb(appuser)

	// ASSERT
	assert.Equal(t, expected, result)

}
func TestChannelService_Create(t *testing.T) {
	t.Run("Successful:create_channel_for_appserver", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)
		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: appserver.ID}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		channel, err := svc.Create(createObj)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
	})

	t.Run("Error:returns_error_fail_create", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedChannel := qx.Channel{}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, fmt.Errorf("error on create"))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.Create(createObj)

		// ASSERT
		assert.Contains(t, err.Error(), "error on create")
	})
}

func TestChannelService_GetById(t *testing.T) {
	t.Run("Successful:returns_a_channel_object", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedChannel := testutil.TestChannel(t, nil)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", ctx, expectedChannel.ID).Return(*expectedChannel, nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		channel, err := svc.GetById(expectedChannel.ID)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
	})

	t.Run("Error:when_no_rows_returned_errors_not_found", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := testutil.TestChannel(t, nil)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(*channel, fmt.Errorf("no rows in result set"))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:on_database_error_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := testutil.TestChannel(t, nil)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(*channel, fmt.Errorf("error on create"))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Contains(t, err.Error(), "error on create")
	})
}

func TestChannelService_List(t *testing.T) {
	t.Run("Successful:with_appserver_filter", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		var serverFilter = pgtype.UUID{Valid: true, Bytes: appserverId}
		expected := []qx.Channel{
			{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()},
			{ID: uuid.New(), Name: "bar", AppserverID: uuid.New()},
		}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: serverFilter}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListServerChannels", ctx, queryParams).Return(expected, nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		result, err := svc.List(queryParams)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, result, expected)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		var serverFilter = pgtype.UUID{Valid: true, Bytes: appserverId}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: serverFilter}
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListServerChannels", ctx, queryParams).Return([]qx.Channel{}, fmt.Errorf("database error"))

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		_, err := svc.List(queryParams)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestChannelService_Delete(t *testing.T) {
	t.Run("Successful:can_delete_channel", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannel", ctx, channelId).Return(int64(1), nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(channelId)

		// ASSERT
		assert.Equal(t, err, nil)
	})

	t.Run("Error:errors_when_no_channel_found", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannel", ctx, channelId).Return(int64(0), nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(channelId)

		// ASSERT
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("Error:when_delete_fails_it_errors", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		channelId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteChannel", ctx, channelId).Return(int64(0), fmt.Errorf("mock error"))

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		err := svc.Delete(channelId)

		// ASSERT
		assert.Contains(t, err.Error(), "mock error")
	})
}
