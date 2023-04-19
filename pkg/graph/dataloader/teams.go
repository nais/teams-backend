package dataloader

import (
	"context"
	"fmt"

	"github.com/graph-gophers/dataloader"
	"github.com/nais/console/pkg/db"
)

type TeamReader struct {
	db db.Database
}

func (r *TeamReader) GetTeams(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	// read all requested teams in a single query
	teamIDs := make([]string, len(keys))
	for ix, key := range keys {
		teamIDs[ix] = key.String()
	}

	// TODO (only fetch teams requested by keys var)
	teams, err := r.db.GetTeams(ctx)
	if err != nil {
		panic(err)
	}

	teamBySlug := map[string]*db.Team{}
	for _, u := range teams {
		teamBySlug[u.Slug.String()] = u
	}

	output := make([]*dataloader.Result, len(keys))
	for index, teamKey := range keys {
		team, ok := teamBySlug[teamKey.String()]
		if ok {
			output[index] = &dataloader.Result{Data: team, Error: nil}
		} else {
			err := fmt.Errorf("team not found %q", teamKey.String())
			output[index] = &dataloader.Result{Data: nil, Error: err}
		}
	}

	return output
}
