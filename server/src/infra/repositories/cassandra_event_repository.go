package repositories

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
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

func (r *CassandraEventRepository) FindAll(userID string) []*domain.Event {
	stmt, names := qb.Select(database.CassandraKeyspace+".events").
		Where(qb.Eq("user_id")).
		Limit(1000).
		ToCql()

	var events []*domain.Event
	if err := r.session.Query(stmt, names).
		Bind(userID).
		Select(&events); err != nil {
		return []*domain.Event{}
	}

	return events
}