package testutil

import (
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

type MockSyncProducer struct {
	mock.Mock
}

// ------ MockProducer ------
func (m *MockProducer) SendMessage(value interface{}, action event.ActionType, appuser []*appuser.Appuser) error {
	args := m.Called(value, action)
	return args.Error(0)
}

// ------ MockSyncProducer ------

func (m *MockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

// Required method
func (m *MockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	args := m.Called(msgs)
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	args := m.Called()
	return args.Get(0).(sarama.ProducerTxnStatusFlag)
}

// Required method
func (m *MockSyncProducer) IsTransactional() bool {
	args := m.Called()
	return args.Bool(0)
}

// Required method
func (m *MockSyncProducer) BeginTxn() error {
	args := m.Called()
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) CommitTxn() error {
	args := m.Called()
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) AbortTxn() error {
	args := m.Called()
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	args := m.Called(offsets, groupId)
	return args.Error(0)
}

// Required method
func (m *MockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	args := m.Called(msg, groupId, metadata)
	return args.Error(0)
}
