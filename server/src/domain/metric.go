package domain

import (
	"time"

	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
)

type Metric struct {
	Id        string
	Name      string
	Service   string
	Value     string
	Unit      string
	Status    vo.MetricStatus
	Timestamp time.Time
	UserId    string
}

func NewMetric(
	id string,
	name string,
	service string,
	value string,
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
