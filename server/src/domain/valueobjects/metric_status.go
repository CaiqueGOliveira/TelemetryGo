package valueobjects

import "errors"

type MetricStatus string

const (
	MetricStatusOk   MetricStatus = "ok"
	MetricStatusWarn MetricStatus = "warn"
	MetricStatusCrit MetricStatus = "crit"
)

func CreateMetricStatus(value string) (MetricStatus, error) {
	status := MetricStatus(value)
	switch status {
	case MetricStatusOk, MetricStatusWarn, MetricStatusCrit:
		return status, nil
	default:
		return "", errors.New("invalid metric status")
	}
}

func (s MetricStatus) String() string {
	return string(s)
}
