package producer

import "github.com/IBM/sarama"

type MessageProducer interface {
	SendMessage(key, value []byte) error
}

type KafkaProducer struct {
	Producer sarama.SyncProducer
	Topic    string
}

func NewKafkaProducer(p sarama.SyncProducer, topic string) *KafkaProducer {
	return &KafkaProducer{Producer: p, Topic: topic}
}

func (kp *KafkaProducer) SendMessage(key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.ByteEncoder(value),
	}

	if key != nil {
		msg.Key = sarama.ByteEncoder(key)
	}

	partition, offset, err := kp.Producer.SendMessage(msg)

	if err != nil {
		return err
	}

	println("Message sent to partition", partition, "at offset", offset)
	return nil
}
