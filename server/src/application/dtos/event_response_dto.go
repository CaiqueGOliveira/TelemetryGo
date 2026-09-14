package dtos

import "github.com/CaiqueGOliveira/TelemetryGo/src/domain"

type EventResponseDto struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	Service   string `json:"service"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

func ToEventResponseDto(event *domain.Event) EventResponseDto {
	return EventResponseDto{
		Id:        event.Id.String(),
		Type:      event.Type,
		Service:   event.Service,
		Message:   event.Message,
		Severity:  event.Severity.String(),
		Timestamp: event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToEventResponseDtos(events []*domain.Event) []EventResponseDto {
	result := make([]EventResponseDto, 0, len(events))
	for _, event := range events {
		result = append(result, ToEventResponseDto(event))
	}
	return result
}
