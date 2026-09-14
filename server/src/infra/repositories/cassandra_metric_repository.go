package repositories

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/google/uuid"
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

func (r *CassandraMetricRepository) List(userID string, filter repository.MetricFilter, limit int, offset int) ([]*domain.Metric, error) {
	builder := qb.Select(database.CassandraKeyspace + ".metrics").
		Where(qb.Eq("user_id"))

	var args []interface{}
	args = append(args, userID)

	if filter.Status != "" {
		builder = builder.Where(qb.Eq("status"))
		args = append(args, filter.Status)
	}
	if filter.Name != "" {
		builder = builder.Where(qb.Eq("name"))
		args = append(args, filter.Name)
	}
	if filter.Service != "" {
		builder = builder.Where(qb.Eq("service"))
		args = append(args, filter.Service)
	}
	if filter.Start != nil {
		builder = builder.Where(qb.Gt("timestamp"))
		args = append(args, *filter.Start)
	}
	if filter.End != nil {
		builder = builder.Where(qb.Lt("timestamp"))
		args = append(args, *filter.End)
	}

	fetch := limit
	if limit <= 0 {
		fetch = 1000
	}
	if offset > 0 {
		fetch += offset
	}
	builder = builder.Limit(uint(fetch))

	stmt, names := builder.ToCql()

	var metrics []*domain.Metric
	if err := r.session.Query(stmt, names).
		Bind(args...).
		Select(&metrics); err != nil {
		return nil, err
	}

	if offset > 0 {
		if offset >= len(metrics) {
			return []*domain.Metric{}, nil
		}
		metrics = metrics[offset:]
	}

	return metrics, nil
}

func (r *CassandraMetricRepository) Delete(userID string, id uuid.UUID) error {
	return qb.Delete(database.CassandraKeyspace + ".metrics").
		Where(qb.Eq("user_id"), qb.Eq("id")).
		Query(r.session).
		Bind(userID, id.String()).
		Exec()
}