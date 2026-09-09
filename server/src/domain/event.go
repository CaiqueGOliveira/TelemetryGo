package domain

import (
	"time"

	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
)

type Event struct {
	Id        string
	Type      string
	Service   string
	Message   string
	Severity  vo.Severity
	Timestamp time.Time
	UserId    string
}

func NewEvent(
	id string,
	eventType string,
	service string,
	message string,
	severity string,
	timestamp time.Time,
) (*Event, error) {
	sev, err := vo.CreateSeverity(severity)
	if err != nil {
		return nil, err
	}

	return &Event{
		Id:        id,
		Type:      eventType,
		Service:   service,
		Message:   message,
		Severity:  sev,
		Timestamp: timestamp,
	}, nil
}
