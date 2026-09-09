package repositories

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

var _ repository.MetricRepository = (*CassandraMetricRepository)(nil)

type CassandraMetricRepository struct {
	session gocqlx.Session
}

func NewCassandraMetricRepository(session gocqlx.Session) *CassandraMetricRepository {
	return &CassandraMetricRepository{session: session}
}

func (r *CassandraMetricRepository) Save(metric *domain.Metric) error {
	return qb.Insert(database.CassandraKeyspace+".metrics").
		Columns("user_id", "id", "name", "service", "value", "unit", "status", "timestamp").
		Query(r.session).
		BindStruct(metric).
		Exec()
}

func (r *CassandraMetricRepository) FindAll(userID string) []*domain.Metric {
	stmt, names := qb.Select(database.CassandraKeyspace+".metrics").
		Where(qb.Eq("user_id")).
		Limit(1000).
		ToCql()

	var metrics []*domain.Metric
	if err := r.session.Query(stmt, names).
		Bind(userID).
		Select(&metrics); err != nil {
		return []*domain.Metric{}
	}

	return metrics
}