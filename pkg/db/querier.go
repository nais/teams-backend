package db

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func (q *Queries) Transaction(ctx context.Context, callback QuerierTransactionFunc) error {
	querier, err := q.begin(ctx)
	if err != nil {
		return err
	}

	defer querier.rollback(ctx)

	err = callback(ctx, querier)
	if err != nil {
		return err
	}

	return querier.commit(ctx)
}

func (q *Queries) begin(ctx context.Context) (*Queries, error) {
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
