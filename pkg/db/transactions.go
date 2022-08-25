package db

import "context"

func (d *database) Transaction(ctx context.Context, fn TransactionFunc) error {
	return d.querier.Transaction(ctx, func(querier Querier) error {
		return fn(ctx, NewDatabase(querier))
	})
}
