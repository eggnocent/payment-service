package kafka

import (
	config2 "payment-service/config"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Kafka struct {
	brokers []string
}

type IKafka interface {
	ProduceMessage(topic string, data []byte) error
}

func NewKafkaProducer(brokers []string) IKafka {
	return &Kafka{
		brokers: brokers,
	}
}

func (k *Kafka) ProduceMessage(topic string, data []byte) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = config2.Config.Kafka.MaxRetry

	producer, err := sarama.NewSyncProducer(k.brokers, config)
	if err != nil {
		logrus.Errorf("Error creating the producer: %s", err)
		return err
	}

	defer func(producer sarama.SyncProducer) {
		err = producer.Close()
		if err != nil {
			logrus.Errorf("Error closing the producer: %s", err)
			return
		}
	}(producer)

	message := &sarama.ProducerMessage{
		Topic:   topic,
		Headers: nil,
		Value:   sarama.ByteEncoder(data),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		logrus.Errorf("Error sending message to topic %s: %s", topic, err)
		return err
	}

	logrus.Infof("Message sent to topic %s at partition %d, offset %d", topic, partition, offset)
	return nil
}
