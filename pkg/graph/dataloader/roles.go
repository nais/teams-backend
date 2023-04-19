package dataloader

import (
	"context"

	"github.com/graph-gophers/dataloader"
	"github.com/nais/console/pkg/db"
)

type UserRoleReader struct {
	db db.Database
}

func (r *UserRoleReader) GetUserRoles(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	// read all requested roles in a single query
	userIDs := make([]string, len(keys))
	for ix, key := range keys {
		userIDs[ix] = key.String()
	}

	userRoles, err := r.db.GetAllUserRoles(ctx)
	if err != nil {
		panic(err)
	}

	userRolesByUserID := map[string][]*db.UserRole{}
	for _, u := range userRoles {
		current := userRolesByUserID[u.UserID.String()]
		userRolesByUserID[u.UserID.String()] = append(current, u)
	}

	output := make([]*dataloader.Result, len(keys))
	for index, userKey := range keys {
		roles := userRolesByUserID[userKey.String()]
		output[index] = &dataloader.Result{Data: roles, Error: nil}
	}

	return output
}
