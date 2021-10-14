package kafka_wrapper

import (
	"github.com/xutils/lib-common/xlog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducerConfig struct {
	Brokers           string `toml:"brokers"`
	Topic             string `toml:"topic"`
	MaxBufferingMaxMs int    `toml:"max_buffering_max_ms"`
}

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(brokers string, bufferingMaxMs int) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      brokers,
		"go.batch.producer":      true,
		"queue.buffering.max.ms": bufferingMaxMs})
	if err != nil {
		return nil, err
	}

	go kafkaProducerResultLoop(p)
	return &KafkaProducer{producer: p}, nil
}

func kafkaProducerResultLoop(p *kafka.Producer) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				xlog.Error("Kafka delivery failed: %v", m.TopicPartition.Error)
			}
		default:
			xlog.Error("Kafka produce ignored event: %s", ev)
		}
	}
}

func (producer *KafkaProducer) SendMessage(topic string, data []byte) {
	xlog.Debug("msg produced||topic=%v||data=%v", topic, string(data))
	producer.producer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          data}
}

func (producer *KafkaProducer) Close() {
	producer.producer.Close()
}
