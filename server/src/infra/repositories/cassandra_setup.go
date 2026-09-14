package repositories

import (
	"github.com/CaiqueGOliveira/TelemetryGo/src/infra/database"
	"github.com/scylladb/gocqlx/v2"
)

func SetupCassandra(session gocqlx.Session) error {
	statements := []string{
		`CREATE KEYSPACE IF NOT EXISTS ` + database.CassandraKeyspace + `
			WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		`CREATE TABLE IF NOT EXISTS ` + database.CassandraKeyspace + `.events (
			user_id    text,
			id         text,
			type       text,
			service    text,
			message    text,
			severity   text,
			timestamp  timestamp,
			PRIMARY KEY (user_id, id)
		)`,
		`CREATE TABLE IF NOT EXISTS ` + database.CassandraKeyspace + `.metrics (
			user_id    text,
			id         text,
			name       text,
			service    text,
			value      double,
			unit       text,
			status     text,
			timestamp  timestamp,
			PRIMARY KEY (user_id, id)
		)`,
	}

	for _, stmt := range statements {
		if err := session.ExecStmt(stmt); err != nil {
			return err
		}
	}

	return nil
}
