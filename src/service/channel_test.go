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
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestChannelService_PgTypeToPb(t *testing.T) {

	// ARRANGE
	ctx := context.Background()
	svc := service.NewChannelService(ctx, testutil.TestDbConn, new(testutil.MockQuerier), new(testutil.MockProducer))

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

	t.Run("Successful:create_channel_for_appserver", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, false)
		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: appserver.ID}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}
		roleFilterParams := qx.FilterChannelRoleParams{ChannelID: pgtype.UUID{Bytes: expectedChannel.ID, Valid: true}}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannelRole", ctx, roleFilterParams).Return([]qx.FilterChannelRoleRow{}, nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, expectedChannel.AppserverID).Return([]qx.ListAppserverUserSubsRow{
			{ID: uuid.New(), AppserverSubID: expectedChannel.AppserverID},
		}, nil)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_ADD_CHANNEL, mock.Anything).Return(nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		channel, err := svc.Create(createObj)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
	})

	t.Run("Successful:error_listing_user_subs_for_notifications_to_users_does_not_impact_result", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, false)
		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: appserver.ID}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}
		roleFilterParams := qx.FilterChannelRoleParams{ChannelID: pgtype.UUID{Bytes: expectedChannel.ID, Valid: true}}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannelRole", ctx, roleFilterParams).Return([]qx.FilterChannelRoleRow{}, nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, expectedChannel.AppserverID).Return(nil, fmt.Errorf("boom"))
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)
		mockProducer.On("NotifyMessageFailure", mock.Anything).Return(nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		channel, err := svc.Create(createObj)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
	})

	t.Run("Successful:error_getting_roles_for_notifications_to_users_does_not_impact_result", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, false)
		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: appserver.ID}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}
		roleFilterParams := qx.FilterChannelRoleParams{ChannelID: pgtype.UUID{Bytes: expectedChannel.ID, Valid: true}}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannelRole", ctx, roleFilterParams).Return(nil, fmt.Errorf("boom"))
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)
		mockProducer.On("NotifyMessageFailure", mock.Anything).Return(nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

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
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(nil, fmt.Errorf("error on create"))
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_ADD_CHANNEL, mock.Anything).Return(nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.Create(createObj)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create channel error: error on create")
	})

	t.Run("Error:on_producer_message_error_it_does_nothing", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil, false)
		expectedChannel := qx.Channel{ID: uuid.New(), Name: "foo", AppserverID: appserver.ID}
		createObj := qx.CreateChannelParams{Name: expectedChannel.Name, AppserverID: expectedChannel.AppserverID}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return([]qx.FilterChannelRoleRow{}, nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, mock.Anything).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("CreateChannel", ctx, createObj).Return(expectedChannel, nil)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_ADD_CHANNEL, mock.Anything).Return(fmt.Errorf("boom"))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		channel, err := svc.Create(createObj)

		// ASSERT
		assert.Nil(t, err)
		assert.Equal(t, expectedChannel.ID, channel.ID)
		assert.Equal(t, expectedChannel.Name, channel.Name)
		assert.Equal(t, expectedChannel.AppserverID, channel.AppserverID)
	})
}

func TestChannelService_GetById(t *testing.T) {

	t.Run("Successful:returns_a_channel_object", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedChannel := testutil.TestChannel(t, nil, false)

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, expectedChannel.ID).Return(*expectedChannel, nil)
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

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
		channel := testutil.TestChannel(t, nil, false)

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(nil, fmt.Errorf(message.DbNotFound))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
	})

	t.Run("Error:on_database_error_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := testutil.TestChannel(t, nil, false)

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, channel.ID).Return(*channel, fmt.Errorf("error get by id"))
		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.GetById(channel.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: error get by id")
	})
}

func TestChannelService_List(t *testing.T) {

	t.Run("Successful:with_appserver_filter", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		expected := []qx.Channel{
			{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()},
			{ID: uuid.New(), Name: "bar", AppserverID: uuid.New()},
		}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("ListServerChannels", ctx, queryParams).Return(expected, nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		result, err := svc.ListServerChannels(queryParams)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, result, expected)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		var nameFilter = pgtype.Text{Valid: false, String: ""}
		queryParams := qx.ListServerChannelsParams{Name: nameFilter, AppserverID: appserverId}
		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("ListServerChannels", ctx, queryParams).Return(nil, fmt.Errorf("database error"))

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.ListServerChannels(queryParams)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: database error")
	})
}

func TestChannelService_Filter(t *testing.T) {

	t.Run("Successful:with_appserver_filter", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		expected := []qx.Channel{
			{ID: uuid.New(), Name: "foo", AppserverID: uuid.New()},
			{ID: uuid.New(), Name: "bar", AppserverID: uuid.New()},
		}
		queryParams := qx.FilterChannelParams{AppserverID: appserverId}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannel", ctx, queryParams).Return(expected, nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		result, err := svc.Filter(queryParams)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, result, expected)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		appserverId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		queryParams := qx.FilterChannelParams{AppserverID: appserverId}
		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("FilterChannel", ctx, queryParams).Return(nil, fmt.Errorf("database error"))

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		_, err := svc.Filter(queryParams)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: database error")
	})
}

func TestChannelService_Delete(t *testing.T) {

	t.Run("Successful:can_delete_public_channel", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		channelRows := []qx.FilterChannelRoleRow{
			{ChannelID: uuid.New(), AppserverID: uuid.New()},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
		mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return(channelRows, nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, mock.Anything).Return([]qx.ListAppserverUserSubsRow{
			{ID: uuid.New(), Username: "testuser"},
		}, nil)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_CHANNEL, mock.Anything).Return(nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err, nil)
		mockProducer.AssertExpectations(t)
	})

	t.Run("Successful:can_delete_private_channel", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New(), IsPrivate: true}
		channelRows := []qx.FilterChannelRoleRow{
			{ChannelID: uuid.New(), AppserverID: uuid.New()},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
		mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return(channelRows, nil)
		mockQuerier.On("GetChannelUsersByRoles", ctx, mock.Anything).Return([]qx.Appuser{
			{ID: uuid.New(), Username: "testuser"},
		}, nil)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_CHANNEL, mock.Anything).Return(nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err, nil)
		mockProducer.AssertExpectations(t)
	})

	t.Run(
		"Successful:error_getting_channel_by_id_for_notifications_does_not_impact_result",
		func(t *testing.T,
		) {
			ctx := testutil.Setup(t, func() {})
			c := qx.Channel{ID: uuid.New()}
			channelRows := []qx.FilterChannelRoleRow{
				{ChannelID: uuid.New(), AppserverID: uuid.New()},
			}
			mockQuerier := new(testutil.MockQuerier)
			mockProducer := new(testutil.MockProducer)

			mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, fmt.Errorf("boom"))
			mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
			mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return(channelRows, nil)

			svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

			// ACT
			err := svc.Delete(c.ID)

			// ASSERT
			assert.Equal(t, err, nil)
		})

	t.Run(
		"Successful:error_getting_channel_by_roles_for_notifications_does_not_impact_result",
		func(t *testing.T,
		) {
			ctx := testutil.Setup(t, func() {})
			c := qx.Channel{ID: uuid.New(), IsPrivate: true}
			channelRows := []qx.FilterChannelRoleRow{
				{ChannelID: uuid.New(), AppserverID: uuid.New()},
			}
			mockQuerier := new(testutil.MockQuerier)
			mockProducer := new(testutil.MockProducer)

			mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
			mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
			mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return(channelRows, nil)
			mockQuerier.On("GetChannelUsersByRoles", ctx, mock.Anything).Return(nil, fmt.Errorf("boom"))

			svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

			// ACT
			err := svc.Delete(c.ID)

			// ASSERT
			assert.Equal(t, err, nil)
			// assert.True(t, false)
		})

	t.Run(
		"Successful:error_getting_users_with_channel_roles_for_notifications_does_not_impact_result",
		func(t *testing.T,
		) {
			ctx := testutil.Setup(t, func() {})
			c := qx.Channel{ID: uuid.New()}
			channelRows := []qx.FilterChannelRoleRow{
				{ChannelID: uuid.New(), AppserverID: uuid.New()},
			}
			mockQuerier := new(testutil.MockQuerier)
			mockProducer := new(testutil.MockProducer)

			mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
			mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(1), nil)
			mockQuerier.On("FilterChannelRole", ctx, mock.Anything).Return(channelRows, nil)
			mockQuerier.On("ListAppserverUserSubs", ctx, mock.Anything).Return(nil, fmt.Errorf("boom"))

			svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)
			// ACT
			err := svc.Delete(c.ID)

			// ASSERT
			assert.Equal(t, err, nil)
		})

	t.Run("Error:errors_when_no_channel_found", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(int64(0), nil)

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
	})

	t.Run("Error:when_delete_fails_it_errors", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		c := qx.Channel{ID: uuid.New()}
		mockQuerier := new(testutil.MockQuerier)
		mockProducer := new(testutil.MockProducer)
		mockQuerier.On("GetChannelById", ctx, c.ID).Return(c, nil)
		mockQuerier.On("DeleteChannel", ctx, c.ID).Return(nil, fmt.Errorf("mock error"))

		svc := service.NewChannelService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(c.ID)

		// ASSERT
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "mock error")
	})
}
