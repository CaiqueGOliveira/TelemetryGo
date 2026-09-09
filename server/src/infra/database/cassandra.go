package database

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

const CassandraKeyspace = "telemetrygo"

func CassandraConnect(hosts []string) (gocqlx.Session, error) {
	if len(hosts) == 0 {
		return gocqlx.Session{}, errors.New("cassandra hosts are required")
	}

	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = ""
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	session, err := cluster.CreateSession()
	if err != nil {
		return gocqlx.Session{}, err
	}

	return gocqlx.WrapSession(session, err)
}
