//go:build integration
// +build integration

package kafka_test

import (
	"context"
	v1 "github.com/jacktantram/user-service/build/go/events/user/v1"
	"github.com/jacktantram/user-service/pkg/driver/v1/kafka"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestSyncProducer(t *testing.T) kafka.SyncProducer {
	t.Helper()

	producer, err := kafka.NewSyncProducer(kafka.ProducerConfig{}, "localhost:29092")
	require.NoError(t, err)
	return producer
}

func TestSyncProducer_ProduceMessage(t *testing.T) {
	t.Parallel()
	retrievedEvent := &v1.UserCreatedEvent{}
	_, offset, err := NewTestSyncProducer(t).ProduceMessage(context.Background(), "test.topic", retrievedEvent)
	require.NoError(t, err)
	assert.NotEqual(t, 0, offset)
}
