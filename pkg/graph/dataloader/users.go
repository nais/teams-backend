package dataloader

import (
	"context"
	"fmt"

	"github.com/graph-gophers/dataloader"
	"github.com/nais/console/pkg/db"
)

type UserReader struct {
	db db.Database
}

func (r *UserReader) GetUsers(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	// read all requested users in a single query
	userIDs := make([]string, len(keys))
	for ix, key := range keys {
		userIDs[ix] = key.String()
	}

	// TODO (only fetch users requested by keys var)
	users, err := r.db.GetUsers(ctx)
	if err != nil {
		panic(err)
	}

	userById := map[string]*db.User{}
	for _, u := range users {
		userById[u.ID.String()] = u
	}

	output := make([]*dataloader.Result, len(keys))
	for index, userKey := range keys {
		user, ok := userById[userKey.String()]
		if ok {
			output[index] = &dataloader.Result{Data: user, Error: nil}
		} else {
			err := fmt.Errorf("user not found %s", userKey.String())
			output[index] = &dataloader.Result{Data: nil, Error: err}
		}
	}

	return output
}
