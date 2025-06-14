package producer_test

import (
	"context"
	"errors"
	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/testutil"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationJob(t *testing.T) {
	t.Run("TestNotificationJob_Execute", func(t *testing.T) {
		t.Run("Success:event_action_add_channel_successfully_sends_message", func(t *testing.T) {
			// ARANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			mockData := &channel.Channel{}
			mockRedis.On("Publish", ctx, "channel", mock.Anything).Return(redis.NewIntCmd(ctx))
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				mockData,
				event.ActionType_ACTION_ADD_CHANNEL,
				nil,
				mockRedis,
			)

			// ACT

			err := notification.Execute(1)

			// ASSERT
			assert.NoError(t, err)
			mockRedis.AssertExpectations(t)
		})

		t.Run("Success:event_action_list_channel_successfully_sends_message", func(t *testing.T) {
			// ARANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			mockData := []*channel.Channel{}
			mockRedis.On("Publish", ctx, "channel", mock.Anything).Return(redis.NewIntCmd(ctx))
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				mockData,
				event.ActionType_ACTION_LIST_CHANNELS,
				nil,
				mockRedis,
			)

			// ACT

			err := notification.Execute(1)

			// ASSERT
			assert.NoError(t, err)
			mockRedis.AssertExpectations(t)
		})

		t.Run("Error:event_action_add_channel_invalid_data_structures_have_marshall_error", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				"boom",
				event.ActionType_ACTION_ADD_CHANNEL,
				nil,
				mockRedis,
			)

			// ACT
			err := notification.Execute(1)

			// ASSERT
			assert.Error(t, err)
			testutil.AssertCustomErrorContains(t, err, "invalid data for action")
			mockRedis.AssertNotCalled(t, "Publish", mock.Anything)
		})

		t.Run("Error:event_action_list_channel_invalid_data_structures_have_marshall_error", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				"boom",
				event.ActionType_ACTION_LIST_CHANNELS,
				nil,
				mockRedis,
			)

			// ACT
			err := notification.Execute(1)

			// ASSERT
			assert.Error(t, err)
			testutil.AssertCustomErrorContains(t, err, "invalid data for action")
			mockRedis.AssertNotCalled(t, "Publish", mock.Anything)
		})
		t.Run("Error:message_not_sent", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			mockData := &channel.Channel{}
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				mockData,
				event.ActionType_ACTION_ADD_CHANNEL,
				nil,
				mockRedis,
			)

			cmd := redis.NewIntCmd(ctx)
			cmd.SetErr(errors.New("message not sent"))

			mockRedis.On("Publish", ctx, "channel", mock.Anything).Return(cmd)

			// ACT
			err := notification.Execute(1)

			// ASSERT
			assert.Error(t, err)
			testutil.AssertCustomErrorContains(t, err, "error sending data to redis: message not sent")
		})
	})

	t.Run("TestNotificationJob_Ctx", func(t *testing.T) {
		t.Run("Success:returns_correct_context", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			mockRedis := new(testutil.MockRedis)
			mockData := &channel.Channel{}
			notification := producer.NewNotificationJob(
				ctx,
				"channel",
				mockData,
				event.ActionType_ACTION_ADD_CHANNEL,
				nil,
				mockRedis,
			)

			// ACT
			resultCtx := notification.Ctx()

			// ASSERT
			assert.Equal(t, ctx, resultCtx)
		})
	})
}

func TestStopWorkerJob(t *testing.T) {
	t.Run("TestStopWorkerJob_Execute", func(t *testing.T) {
		t.Run(("Success:it_does_nothing_and_returns_nil_error"), func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			stopJob := producer.NewStopWorkerJob(ctx)

			// ACT
			err := stopJob.Execute(1)

			// ASSERT
			assert.NoError(t, err)
		})
	})

	t.Run("TestStopWorkerJob_Ctx", func(t *testing.T) {
		t.Run("Success:returns_correct_context", func(t *testing.T) {
			// ARRANGE
			ctx := context.Background()
			stopJob := producer.NewStopWorkerJob(ctx)

			// ACT
			resultCtx := stopJob.Ctx()

			// ASSERT
			assert.Equal(t, ctx, resultCtx)
		})
	})
}
