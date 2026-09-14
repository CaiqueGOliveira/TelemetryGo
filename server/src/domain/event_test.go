package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewEventValidSeverity(t *testing.T) {
	event, err := NewEvent(uuid.New(), "deploy", "api-gateway", "something broke", "critical", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != "deploy" {
		t.Errorf("expected type deploy, got %s", event.Type)
	}
	if event.Service != "api-gateway" {
		t.Errorf("expected service api-gateway, got %s", event.Service)
	}
	if event.Message != "something broke" {
		t.Errorf("unexpected message %s", event.Message)
	}
	if event.Severity.String() != "critical" {
		t.Errorf("expected severity critical, got %s", event.Severity.String())
	}
}

func TestNewEventInvalidSeverity(t *testing.T) {
	if _, err := NewEvent(uuid.New(), "deploy", "api", "msg", "exploded", time.Now()); err == nil {
		t.Fatal("expected error for invalid severity")
	}
}

func TestNewEventEmptySeverity(t *testing.T) {
	if _, err := NewEvent(uuid.New(), "deploy", "api", "msg", "", time.Now()); err == nil {
		t.Fatal("expected error for empty severity")
	}
}
