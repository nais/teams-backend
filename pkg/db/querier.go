package db

import (
	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/sqlc"
)

type Querier interface {
	sqlc.Querier
	WithTx(tx pgx.Tx) Querier
	Conn() *pgx.Conn
}

type Queries struct {
	*sqlc.Queries
	conn *pgx.Conn
}

func Wrap(q *sqlc.Queries, conn *pgx.Conn) Querier {
	return &Queries{
		Queries: q,
		conn:    conn,
	}
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
	}
}

func (q *Queries) Conn() *pgx.Conn {
	return q.conn
}
