package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

type Publisher interface {
	Publish(context.Context, string, []byte, []byte) error
}

type Consumer interface {
	Poll(context.Context, func(key, value []byte) error) error
	Close()
}

type Client struct {
	client *kgo.Client
}

func NewClient(group string, topics ...string) (*Client, error) {
	brokers := strings.Split(valueOrDefault(os.Getenv("KAFKA_BROKERS"), "redpanda:9092"), ",")
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.NoCompression()),
	}
	if len(topics) > 0 {
		opts = append(opts, kgo.ConsumeTopics(topics...))
	}
	if group != "" {
		opts = append(opts, kgo.ConsumerGroup(group))
	}
	if conn := os.Getenv("EVENTHUB_CONNECTION_STRING"); conn != "" {
		opts = append(opts,
			kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
			kgo.SASL(plain.Auth{User: "$ConnectionString", Pass: conn}.AsMechanism()),
		)
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) Publish(ctx context.Context, topic string, key, value []byte) error {
	return c.client.ProduceSync(ctx, &kgo.Record{Topic: topic, Key: key, Value: value}).FirstErr()
}

func (c *Client) Poll(ctx context.Context, handle func(key, value []byte) error) error {
	fetches := c.client.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		return errors.New(errs[0].Err.Error())
	}
	var handlerErr error
	fetches.EachRecord(func(record *kgo.Record) {
		if handlerErr != nil {
			return
		}
		handlerErr = handle(record.Key, record.Value)
	})
	if handlerErr != nil {
		return handlerErr
	}
	c.client.CommitUncommittedOffsets(ctx)
	return nil
}

func (c *Client) Close() {
	c.client.CloseAllowingRebalance()
}

func Topic() string {
	return valueOrDefault(os.Getenv("KAFKA_TOPIC"), "inventory.events")
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func Backoff(attempts int) time.Duration {
	delay := time.Duration(1<<min(attempts, 6)) * time.Second
	return min(delay, 30*time.Second)
}
