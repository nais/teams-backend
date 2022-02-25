package graph

import (
	"context"
	"fmt"
)

func (r *Resolver) updateOrBust(ctx context.Context, obj interface{}) error {
	tx := r.db.WithContext(ctx).Updates(obj)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no such %T", obj)
	}
	tx.Find(obj)
	return tx.Error
}
