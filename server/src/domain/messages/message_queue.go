package messages

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, func(), error)
}
