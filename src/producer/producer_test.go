package producer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/testutil"
)

type redisBaseCmd struct {
	err error
}

func TestNewMProducer(t *testing.T) {
	mockRedis := new(testutil.MockRedis)

	mp := producer.NewMProducer(mockRedis)

	assert.NotNil(t, mp)
	assert.Equal(t, mockRedis, mp.Redis)
}

func TestMProducer_SendMessage(t *testing.T) {
	t.Run("Success:event_action_add_channel_successfully_sends_message", func(t *testing.T) {
		// ARANGE
		ctx := context.Background()
		mockRedis := new(testutil.MockRedis)
		mockData := &channel.Channel{}
		mockRedis.On("Publish", ctx, "channel", mock.Anything).Return(redis.NewIntCmd(ctx))

		kp := &producer.MProducer{Redis: mockRedis}
		// ACT
		err := kp.SendMessage(ctx, "channel", mockData, event.ActionType_ACTION_ADD_CHANNEL, nil)

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

		kp := &producer.MProducer{Redis: mockRedis}
		// ACT
		err := kp.SendMessage(ctx, "channel", mockData, event.ActionType_ACTION_LIST_CHANNELS, nil)

		// ASSERT
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:event_action_add_channel_invalid_data_structures_have_marshall_error", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		mockRedis := new(testutil.MockRedis)
		kp := &producer.MProducer{Redis: mockRedis}

		// ACT
		err := kp.SendMessage(ctx, "channel", "boom", event.ActionType_ACTION_ADD_CHANNEL, nil)

		// ASSERT
		assert.Error(t, err)
		testutil.AssertCustomErrorContains(t, err, "invalid data for action")
		mockRedis.AssertNotCalled(t, "Publish", mock.Anything)
	})

	t.Run("Error:event_action_list_channel_invalid_data_structures_have_marshall_error", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		mockRedis := new(testutil.MockRedis)
		kp := &producer.MProducer{Redis: mockRedis}

		// ACT
		err := kp.SendMessage(ctx, "channel", "boom", event.ActionType_ACTION_LIST_CHANNELS, nil)

		// ASSERT
		assert.Error(t, err)
		testutil.AssertCustomErrorContains(t, err, "invalid data for action")
		mockRedis.AssertNotCalled(t, "Publish", mock.Anything)
	})
	t.Run("Error:message_not_sent", func(t *testing.T) {
		// ARRANGE
		ctx := context.Background()
		mockRedis := new(testutil.MockRedis)
		kp := &producer.MProducer{Redis: mockRedis}

		mockData := &channel.Channel{}

		cmd := redis.NewIntCmd(ctx)
		cmd.SetErr(errors.New("message not sent"))

		mockRedis.On("Publish", ctx, "channel", mock.Anything).Return(cmd)

		// ACT
		err := kp.SendMessage(ctx, "channel", mockData, event.ActionType_ACTION_ADD_CHANNEL, nil)

		// ASSERT
		assert.Error(t, err)
		testutil.AssertCustomErrorContains(t, err, "error sending data to redis: message not sent")
	})
}
