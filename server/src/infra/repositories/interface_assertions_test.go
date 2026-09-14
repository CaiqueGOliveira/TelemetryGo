package repositories

import (
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
)

var (
	_ r.UserRepository   = (*UserRepository)(nil)
	_ r.EventRepository  = (*InMemoryEventRepository)(nil)
	_ r.MetricRepository = (*InMemoryMetricRepository)(nil)
	_ r.UserRepository   = (*PostgresUserRepository)(nil)
	_ r.EventRepository  = (*CassandraEventRepository)(nil)
	_ r.MetricRepository = (*CassandraMetricRepository)(nil)
)
