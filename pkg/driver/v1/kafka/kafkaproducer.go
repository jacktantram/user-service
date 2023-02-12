package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"google.golang.org/protobuf/proto"
)

// SyncProducer is responsible for writing messages to a particular topic
type SyncProducer struct {
	p sarama.SyncProducer
}

type ProducerConfig struct {
}

// NewSyncProducer creates a new synchronous producer
func NewSyncProducer(p ProducerConfig, hosts ...string) (SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(hosts, config)
	if err != nil {
		return SyncProducer{}, err
	}
	return SyncProducer{p: producer}, err
}

// ProduceMessage provides functionality for writing a proto message to a topic
// Could be improved by adding restrictions on topic name, i.e. $domain.$entity-$action_v$version
func (p SyncProducer) ProduceMessage(ctx context.Context, topic string, msg proto.Message) (partition int32, offset int64, err error) {
	protoBytes, err := proto.Marshal(msg)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to marshal proto bytes: %w", err)
	}
	partition, offset, err = p.p.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(protoBytes),
	})
	if err != nil {
		return partition, offset, err
	}
	return partition, offset, nil
}
