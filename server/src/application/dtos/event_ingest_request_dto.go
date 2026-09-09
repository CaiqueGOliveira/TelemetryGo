package dtos

type EventIngestRequestDto struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp,omitempty"`
}

func NewEventIngestRequestDto(id string, eventType string, service string, message string, severity string, timestamp string) *EventIngestRequestDto {
	return &EventIngestRequestDto{
		ID:        id,
		Type:      eventType,
		Service:   service,
		Message:   message,
		Severity:  severity,
		Timestamp: timestamp,
	}
}
