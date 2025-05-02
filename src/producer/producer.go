package producer

import (
	"fmt"
	"mist/src/errors/message"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type MessageProducer interface {
	SendMessage(interface{}, event.ActionType) error
}

type KafkaProducer struct {
	Producer sarama.SyncProducer
	Topic    string
}

func NewKafkaProducer(p sarama.SyncProducer, topic string) *KafkaProducer {
	return &KafkaProducer{Producer: p, Topic: topic}
}

func (kp *KafkaProducer) SendMessage(data interface{}, action event.ActionType) error {
	e, err := kp.marshall(data, action)

	if err != nil {
		return message.UnknownError(fmt.Sprintf("error marshalling data for kafka: %v", err))
	}

	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.ByteEncoder(e),
	}
	_, _, err = kp.Producer.SendMessage(msg)

	if err != nil {
		return message.UnknownError(fmt.Sprintf("error sending data to kafka: %v", err))
	}

	println("Message to kafka successfully sent")
	return nil
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

func (kp *KafkaProducer) marshall(data interface{}, action event.ActionType) ([]byte, error) {
	var e *event.Event
	switch action {

	case event.ActionType_ACTION_CREATE_CHANNEL:
		d, ok := data.(*channel.Channel)
		if !ok {
			return nil, fmt.Errorf("invalid data type for action %v", action)
		}

		data = &event.Event{
			Meta: &event.Meta{Action: action},
			Data: &event.Event_CreateChannel{
				CreateChannel: &event.CreateChannel{
					Channel: d,
				},
			},
		}

	}
	return proto.Marshal(e)
}
