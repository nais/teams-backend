package db

import (
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/sqlc"
)

type Querier interface {
	sqlc.Querier
	WithTx(tx pgx.Tx) Querier
	Conn() *pgxpool.Pool
}

type Queries struct {
	*sqlc.Queries
	connPool *pgxpool.Pool
}

func Wrap(q *sqlc.Queries, connPool *pgxpool.Pool) Querier {
	return &Queries{
		Queries:  q,
		connPool: connPool,
	}
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
	}
}

func (q *Queries) Conn() *pgxpool.Pool {
	return q.connPool
}
