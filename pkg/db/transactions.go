package db

import "context"

func (d *database) Transaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := d.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = fn(ctx, NewDatabase(d.querier.WithTx(tx), d.connPool))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
