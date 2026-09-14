package repository

import "time"

type EventFilter struct {
	Severity string
	Type     string
	Service  string
	Start    *time.Time
	End      *time.Time
}

type MetricFilter struct {
	Status  string
	Name    string
	Service string
	Start   *time.Time
	End     *time.Time
}