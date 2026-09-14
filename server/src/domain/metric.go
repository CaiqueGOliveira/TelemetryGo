package domain

import (
	"time"

	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
	"github.com/google/uuid"
)

type Metric struct {
	Id        uuid.UUID
	Name      string
	Service   string
	Value     float64
	Unit      string
	Status    vo.MetricStatus
	Timestamp time.Time
	UserId    string
}

func NewMetric(
	id uuid.UUID,
	name string,
	service string,
	value float64,
	unit string,
	status string,
	timestamp time.Time,
) (*Metric, error) {
	metricStatus, err := vo.CreateMetricStatus(status)
	if err != nil {
		return nil, err
	}

	return &Metric{
		Id:        id,
		Name:      name,
		Service:   service,
		Value:     value,
		Unit:      unit,
		Status:    metricStatus,
		Timestamp: timestamp,
	}, nil
}
