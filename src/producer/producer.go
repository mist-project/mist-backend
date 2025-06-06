package producer

import (
	"fmt"
	"log/slog"
	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type MessageProducer interface {
	SendMessage(interface{}, event.ActionType, []*appuser.Appuser) error
	NotifyMessageFailure(error) error
}

type KafkaProducer struct {
	Producer sarama.SyncProducer
	Topic    string
}

func NewKafkaProducer(p sarama.SyncProducer, topic string) *KafkaProducer {
	return &KafkaProducer{Producer: p, Topic: topic}
}

func (kp *KafkaProducer) SendMessage(data interface{}, action event.ActionType, appusers []*appuser.Appuser) error {
	e, err := kp.marshall(data, action, appusers)

	if err != nil {
		return faults.ExtendError(err)
	}

	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.ByteEncoder(e),
	}
	_, _, err = kp.Producer.SendMessage(msg)

	if err != nil {
		return faults.MessageProducerError(fmt.Sprintf("error sending data to kafka: %v", err), slog.LevelError)
	}

	return err
}

// func (kp *KafkaProducer) SendMessageWithKey(key, value []byte) error {
// 	msg := &sarama.ProducerMessage{
// 		Topic: kp.Topic,
// 		Value: sarama.ByteEncoder(value),
// 	}

// 	if key != nil {
// 		msg.Key = sarama.ByteEncoder(key)
// 	}

// 	partition, offset, err := kp.Producer.SendMessage(msg)

// 	if err != nil {
// 		return err
// 	}

// 	println("Message sent to partition", partition, "at offset", offset)
// 	return nil
// }

func (kp *KafkaProducer) marshall(data interface{}, action event.ActionType, appusers []*appuser.Appuser) ([]byte, error) {
	var e *event.Event

	if appusers == nil {
		appusers = []*appuser.Appuser{}
	}

	switch action {
	case event.ActionType_ACTION_ADD_CHANNEL:
		d, ok := data.(*channel.Channel)

		if !ok {
			return nil, faults.MarshallError(fmt.Sprintf("invalid data type for action %v", action), slog.LevelWarn)
		}

		e = &event.Event{
			Meta: &event.Meta{Action: action, Appusers: appusers},
			Data: &event.Event_AddChannel{
				AddChannel: &event.AddChannel{
					Channel: d,
				},
			},
		}
	}

	return proto.Marshal(e)
}

func (kp *KafkaProducer) NotifyMessageFailure(err error) error {
	return message.UnknownError(fmt.Sprintf("error notifying message failure to kafka: %v", err))
}
