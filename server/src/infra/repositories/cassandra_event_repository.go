package repositories

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

var _ repository.EventRepository = (*CassandraEventRepository)(nil)

type CassandraEventRepository struct {
	session gocqlx.Session
}

func NewCassandraEventRepository(session gocqlx.Session) *CassandraEventRepository {
	return &CassandraEventRepository{session: session}
}

func (r *CassandraEventRepository) Save(event *domain.Event) error {
	return qb.Insert(database.CassandraKeyspace+".events").
		Columns("user_id", "id", "type", "service", "message", "severity", "timestamp").
		Query(r.session).
		BindStruct(event).
		Exec()
}

func (r *CassandraEventRepository) List(userID string, filter repository.EventFilter, limit int, offset int) ([]*domain.Event, error) {
	builder := qb.Select(database.CassandraKeyspace + ".events").
		Where(qb.Eq("user_id"))

	var args []interface{}
	args = append(args, userID)

	if filter.Severity != "" {
		builder = builder.Where(qb.Eq("severity"))
		args = append(args, filter.Severity)
	}
	if filter.Type != "" {
		builder = builder.Where(qb.Eq("type"))
		args = append(args, filter.Type)
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

	var events []*domain.Event
	if err := r.session.Query(stmt, names).
		Bind(args...).
		Select(&events); err != nil {
		return nil, err
	}

	if offset > 0 {
		if offset >= len(events) {
			return []*domain.Event{}, nil
		}
		events = events[offset:]
	}

	return events, nil
}

func (r *CassandraEventRepository) Delete(userID string, id uuid.UUID) error {
	return qb.Delete(database.CassandraKeyspace + ".events").
		Where(qb.Eq("user_id"), qb.Eq("id")).
		Query(r.session).
		Bind(userID, id.String()).
		Exec()
}