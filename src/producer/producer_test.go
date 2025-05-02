package producer_test

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/producer"
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

		mockProducer.On("SendMessage", mock.MatchedBy(func(msg *sarama.ProducerMessage) bool {
			// Optional: deep check of fields if you want
			return msg.Topic == "test-topic"
		})).Return(int32(1), int64(42), nil)

		// ACT
		err := kp.SendMessage([]byte("key"), []byte("value"))

		// ASSERT
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("Error:message_not_sent", func(t *testing.T) {
		// ARRANGE
		mockProducer := new(testutil.MockSyncProducer)
		kp := &producer.KafkaProducer{
			Producer: mockProducer,
			Topic:    "test-topic",
		}

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), sarama.ErrOutOfBrokers)

		// ACT
		err := kp.SendMessage([]byte("key"), []byte("value"))

		// ASSERT
		assert.Error(t, err)
		mockProducer.AssertExpectations(t)
	})
}
