package messaging

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (r *RedisPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}

func (r *RedisPublisher) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	sub := r.client.Subscribe(ctx, channel)

	ch := make(chan []byte, 64)

	go func() {
		defer close(ch)
		for msg := range sub.Channel() {
			ch <- []byte(msg.Payload)
		}
	}()

	cleanup := func() {
		sub.Unsubscribe(ctx)
		sub.Close()
	}

	return ch, cleanup, nil
}

type NoOpPublisher struct{}

func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

func (p *NoOpPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	return nil
}

func (p *NoOpPublisher) Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error) {
	ch := make(chan []byte)
	cleanup := func() { close(ch) }
	return ch, cleanup, nil
}
