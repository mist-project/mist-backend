package producer_test

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/testutil"
)

func TestNewKafkaProducer(t *testing.T) {
	mockProducer := new(testutil.MockSyncProducer)
	topic := "test-topic"

	kp := producer.NewKafkaProducer(mockProducer, topic)

	assert.NotNil(t, kp)
	assert.Equal(t, topic, kp.Topic)
	assert.Equal(t, mockProducer, kp.Producer)
}

func TestKafkaProducer_SendMessage(t *testing.T) {
	t.Run("Success:message_is_sent", func(t *testing.T) {
		// ARANGE
		mockProducer := new(testutil.MockSyncProducer)
		kp := &producer.KafkaProducer{
			Producer: mockProducer,
			Topic:    "test-topic",
		}
		mockAction := event.ActionType_ACTION_ADD_CHANNEL
		mockData := &channel.Channel{}
		mockProducer.On("SendMessage", mock.Anything).Return(int32(1), int64(42), nil)
		// ACT
		err := kp.SendMessage(mockData, mockAction, nil)

		// ASSERT
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("Error:invalid_data_structures_have_marshall_error", func(t *testing.T) {
		// ARRANGE
		mockProducer := new(testutil.MockSyncProducer)
		kp := &producer.KafkaProducer{
			Producer: mockProducer,
			Topic:    "test-topic",
		}

		mockAction := event.ActionType_ACTION_ADD_CHANNEL

		// ACT
		err := kp.SendMessage("booM", mockAction, nil)

		// ASSERT
		assert.Error(t, err)
		mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything)
	})

	t.Run("Error:message_not_sent", func(t *testing.T) {
		// ARRANGE
		mockProducer := new(testutil.MockSyncProducer)
		kp := &producer.KafkaProducer{
			Producer: mockProducer,
			Topic:    "test-topic",
		}
		mockData := &channel.Channel{}

		mockAction := event.ActionType_ACTION_ADD_CHANNEL
		mockProducer.On("SendMessage", mock.Anything).Return(int32(1), int64(42), sarama.ErrOutOfBrokers)

		// ACT
		err := kp.SendMessage(mockData, mockAction, nil)

		// ASSERT
		assert.Error(t, err)
		mockProducer.AssertExpectations(t)
	})
}

func TestKafkaProducer_NotifyMessageFailure(t *testing.T) {
	// ARRANGE
	errMessage := "error notifying message failure to kafka"
	mockProducer := new(testutil.MockSyncProducer)
	kp := &producer.KafkaProducer{
		Producer: mockProducer,
		Topic:    "test-topic",
	}

	// ACT
	err := kp.NotifyMessageFailure(errors.New(errMessage))

	// ASSERT

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errMessage)
}
