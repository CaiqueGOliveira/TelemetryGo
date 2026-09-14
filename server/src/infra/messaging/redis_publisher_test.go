package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestRedisPublisher(t *testing.T) (*RedisPublisher, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return NewRedisPublisher(client), mr
}

func TestRedisPublisherPublishSubscribe(t *testing.T) {
	pub, _ := newTestRedisPublisher(t)
	ctx := context.Background()

	ch, cleanup, err := pub.Subscribe(ctx, "telemetry:events")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	t.Cleanup(cleanup)

	payload := uuid.NewString()
	if err := pub.Publish(ctx, "telemetry:events", []byte(payload)); err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}

	select {
	case got := <-ch:
		if string(got) != payload {
			t.Errorf("expected %q, got %q", payload, string(got))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestRedisPublisherSubscribeCleanup(t *testing.T) {
	pub, _ := newTestRedisPublisher(t)
	ctx := context.Background()

	ch, cleanup, err := pub.Subscribe(ctx, "telemetry:metrics")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}

	cleanup()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed after cleanup")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected channel to close after cleanup")
	}
}
