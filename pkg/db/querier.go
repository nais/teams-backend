package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/sqlc"
)

type QuerierTxFn func(querier Querier) error

type Querier interface {
	sqlc.Querier
	Transaction(ctx context.Context, callback QuerierTxFn) error
}

type Queries struct {
	*sqlc.Queries
	connPool *pgxpool.Pool
	tx       pgx.Tx
}

func Wrap(q *sqlc.Queries, connPool *pgxpool.Pool) Querier {
	return &Queries{
		Queries:  q,
		connPool: connPool,
	}
}

func (q *Queries) Transaction(ctx context.Context, callback QuerierTxFn) error {
	querier, err := q.begin(ctx)
	if err != nil {
		return err
	}

	defer q.rollback(ctx)

	err = callback(querier)
	if err != nil {
		return err
	}

	return q.commit(ctx)
}

func (q *Queries) begin(ctx context.Context) (Querier, error) {
	var err error
	var tx pgx.Tx

	if q.tx == nil {
		tx, err = q.connPool.Begin(ctx)
	} else {
		tx, err = q.tx.Begin(ctx)
	}

	if err != nil {
		return nil, err
	}

	return &Queries{
		Queries:  q.Queries.WithTx(tx),
		connPool: q.connPool,
		tx:       tx,
	}, nil
}

func (q *Queries) commit(ctx context.Context) error {
	return q.tx.Commit(ctx)
}

func (q *Queries) rollback(ctx context.Context) error {
	return q.tx.Commit(ctx)
}
