package producer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"mist/src/producer"
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

func TestNewMProducerOptions(t *testing.T) {
	mockRedis := new(testutil.MockRedis)

	mp := producer.NewMProducerOptions(mockRedis, &producer.MProducerOptions{
		Workers:     4,
		ChannelSize: 100,
	})

	assert.NotNil(t, mp)
	assert.Equal(t, mockRedis, mp.Redis)
	assert.NotNil(t, 4, mp.Wp)
}

func TestMProducer_SendMessage(t *testing.T) {
	t.Run("Success:it_adds_a_job_to_the_channel", func(t *testing.T) {
		// ARRANGE
		mockRedis := new(testutil.MockRedis)
		mp := producer.NewMProducerOptions(mockRedis, &producer.MProducerOptions{
			Workers:     4,
			ChannelSize: 100,
		})

		// ACT
		mp.SendMessage(
			context.Background(),
			"test_channel",
			"test_data",
			event.ActionType_ACTION_ADD_CHANNEL,
			nil,
		)

		// ASSERT
		assert.Equal(t, mp.Wp.GetJobQueueSize(), 1, "Expected job queue size to be 1")
})
}
