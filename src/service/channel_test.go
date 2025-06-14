package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestChannelService_PgTypeToPb(t *testing.T) {

	sharedDeps := &service.ServiceDeps{
		Db:        new(testutil.MockQuerier),
		MProducer: producer.NewMProducer(new(testutil.MockRedis)),
	}

	// ARRANGE
	ctx := context.Background()
	svc := service.NewChannelService(ctx, sharedDeps)

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

	expected := &channel.Channel{
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

	t.Run("Success:create_channel_for_appserver", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On(
			"ListAppserverUserSubs", ctx, expectedChannel.AppserverID,
		).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		channel, err := svc.Create(createObj)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:returns_error_fail_create", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		expectedChannel := qx.Channel{}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateChannel", ctx, createObj).Return(nil, fmt.Errorf("error on create"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.Create(createObj)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create channel error: error on create")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelService_GetById(t *testing.T) {

	t.Run("Success:returns_a_channel_object", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		expectedChannel := qx.Channel{}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, expectedChannel.ID).Return(expectedChannel, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		channel, err := svc.GetById(expectedChannel.ID)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:when_no_rows_returned_errors_not_found", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_database_error_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		channel := qx.Channel{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(channel, fmt.Errorf("error get by id"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: error get by id")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelService_List(t *testing.T) {

	t.Run("Success:with_appserver_filter", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		expected := []qx.Channel{
			{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()},
			{ID: uuid.New(), Name: "bar", AppserverID: uuid.New()},
		}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListServerChannels", ctx, queryParams).Return(expected, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		result, err := svc.ListServerChannels(queryParams)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, result, expected)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListServerChannels", ctx, queryParams).Return(nil, fmt.Errorf("database error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.ListServerChannels(queryParams)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: database error")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelService_Filter(t *testing.T) {

	t.Run("Success:with_appserver_filter", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		expected := []qx.Channel{
			{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()},
			{ID: uuid.New(), Name: "bar", AppserverID: uuid.New()},
		}
		queryParams := qx.FilterChannelParams{AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("FilterChannel", ctx, queryParams).Return(expected, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		result, err := svc.Filter(queryParams)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, result, expected)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		queryParams := qx.FilterChannelParams{AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("FilterChannel", ctx, queryParams).Return(nil, fmt.Errorf("database error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		_, err := svc.Filter(queryParams)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: database error")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelService_Delete(t *testing.T) {

	t.Run("Success:can_delete_public_channel", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New(), IsPrivate: false, AppserverID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, c.AppserverID).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err, nil)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Success:can_delete_private_channel", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New(), IsPrivate: true, AppserverID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, c.AppserverID).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err, nil)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:errors_when_no_channel_found", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(0), nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:when_get_channel_by_id_fails_it_errors", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, c.ID).Return(nil, fmt.Errorf("mock error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "mock error")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:when_delete_fails_it_errors", func(t *testing.T) {
		ctx, _ := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(nil, fmt.Errorf("mock error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "mock error")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

func TestChannelService_SendChannelListingUpdateNotificationToUsers(t *testing.T) {
	t.Run("Success:sends_channels_for_each_user", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		user1 := qx.ListAppserverUserSubsRow{AppuserID: uuid.New()}
		user2 := qx.ListAppserverUserSubsRow{AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)
		producer.Wp.StartWorkers()
		mockQuerier.On(
			"ListAppserverUserSubs", ctx, appserverID,
		).Return([]qx.ListAppserverUserSubsRow{user1, user2}, nil)

		channel1 := qx.GetChannelsForUsersRow{
			AppuserID:          user1.AppuserID,
			ChannelID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			ChannelName:        pgtype.Text{String: "chan-1", Valid: true},
			ChannelAppserverID: pgtype.UUID{Bytes: appserverID, Valid: true},
			ChannelIsPrivate:   pgtype.Bool{Bool: false, Valid: true},
		}

		channel2 := qx.GetChannelsForUsersRow{
			AppuserID:          user2.AppuserID,
			ChannelID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			ChannelName:        pgtype.Text{String: "chan-2", Valid: true},
			ChannelAppserverID: pgtype.UUID{Bytes: appserverID, Valid: true},
			ChannelIsPrivate:   pgtype.Bool{Bool: true, Valid: true},
		}

		mockQuerier.On(
			"GetChannelsForUsers", ctx,
			qx.GetChannelsForUsersParams{Column1: []uuid.UUID{user1.AppuserID, user2.AppuserID}, AppserverID: appserverID},
		).Return([]qx.GetChannelsForUsersRow{channel1, channel2}, nil)

		mockRedis.On(
			"Publish", ctx, os.Getenv("REDIS_NOTIFICATION_CHANNEL"), mock.Anything,
		).Return(redis.NewIntCmd(ctx)).Once()

		mockRedis.On(
			"Publish", ctx, os.Getenv("REDIS_NOTIFICATION_CHANNEL"), mock.Anything,
		).Return(redis.NewIntCmd(ctx)).Once()

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(nil, appserverID)

		// ASSERT
		producer.Wp.Stop() // Stop the worker pool to ensure all jobs are processed
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:early_return_if_no_users", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverID).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(nil, appserverID)

		// ASSERT
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error:get_channels_fails_logs_and_exits", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		user := qx.ListAppserverUserSubsRow{AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverID).Return([]qx.ListAppserverUserSubsRow{user}, nil)

		mockQuerier.On(
			"GetChannelsForUsers", ctx,
			qx.GetChannelsForUsersParams{Column1: []uuid.UUID{user.AppuserID}, AppserverID: appserverID},
		).Return(nil, fmt.Errorf("db error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(nil, appserverID)

		// ASSERT
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error:get_appserver_user_subs_fails_logs_and_exits", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverID).Return(nil, fmt.Errorf("db error"))

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(nil, appserverID)

		// ASSERT
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Success:sends_channels_for_single_user", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		user := &qx.Appuser{ID: uuid.New()}

		channelRow := qx.GetChannelsForUsersRow{
			AppuserID:          user.ID,
			ChannelID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			ChannelName:        pgtype.Text{String: "chan-1", Valid: true},
			ChannelAppserverID: pgtype.UUID{Bytes: appserverID, Valid: true},
			ChannelIsPrivate:   pgtype.Bool{Bool: true, Valid: true},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)
		producer.Wp.StartWorkers()

		mockQuerier.On(
			"GetChannelsForUsers", ctx,
			qx.GetChannelsForUsersParams{Column1: []uuid.UUID{user.ID}, AppserverID: appserverID},
		).Return([]qx.GetChannelsForUsersRow{channelRow}, nil)

		mockRedis.On(
			"Publish", ctx, os.Getenv("REDIS_NOTIFICATION_CHANNEL"), mock.Anything,
		).Return(redis.NewIntCmd(ctx)).Once()

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(user, appserverID)

		// ASSERT
		producer.Wp.Stop() // Stop the worker pool to ensure all jobs are processed
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("EarlyExit:no_channels_for_users", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()
		user := qx.ListAppserverUserSubsRow{AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverID).Return([]qx.ListAppserverUserSubsRow{user}, nil)

		mockQuerier.On(
			"GetChannelsForUsers", ctx,
			qx.GetChannelsForUsersParams{Column1: []uuid.UUID{user.AppuserID}, AppserverID: appserverID},
		).Return([]qx.GetChannelsForUsersRow{}, nil)

		svc := service.NewChannelService(
			ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer},
		)

		// ACT
		svc.SendChannelListingUpdateNotificationToUsers(nil, appserverID)

		// ASSERT
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertNotCalled(t, "SendMessage", mock.Anything, mock.Anything, mock.Anything)
	})

}
