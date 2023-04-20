package dataloader

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/graph-gophers/dataloader"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
	log "github.com/sirupsen/logrus"
)

type ctxKey string

const loadersKey = ctxKey("dataloaders")

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	Cache           dataloader.Cache
	UserLoader      *dataloader.Loader
	TeamLoader      *dataloader.Loader
	UserRolesLoader *dataloader.Loader
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(database db.Database) *Loaders {
	// define the data loader
	userReader := &UserReader{db: database}
	teamReader := &TeamReader{db: database}
	userRoleReader := &UserRoleReader{db: database}

	cache := dataloader.NewCache()

	opts := []dataloader.Option{
		dataloader.WithCache(cache),
		dataloader.WithInputCapacity(5000),
	}

	loaders := &Loaders{
		Cache:           cache,
		UserLoader:      dataloader.NewBatchedLoader(userReader.GetUsers, opts...),
		TeamLoader:      dataloader.NewBatchedLoader(teamReader.GetTeams, opts...),
		UserRolesLoader: dataloader.NewBatchedLoader(userRoleReader.GetUserRoles, opts...),
	}

	return loaders
}

// Middleware injects data loaders into the context
func Middleware(loaders *Loaders) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCtx := context.WithValue(r.Context(), loadersKey, loaders)
			r = r.WithContext(nextCtx)
			next.ServeHTTP(w, r)

			// clear cache after request is complete
			loaders.Cache.Clear()
			log.Debugf("cleared dataloader cache after request")
		})
	}
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func GetUser(ctx context.Context, userID *uuid.UUID) (*db.User, error) {
	loaders := For(ctx)
	thunk := loaders.UserLoader.Load(ctx, dataloader.StringKey(userID.String()))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*db.User), nil
}

func GetTeam(ctx context.Context, teamSlug *slug.Slug) (*db.Team, error) {
	loaders := For(ctx)
	thunk := loaders.TeamLoader.Load(ctx, dataloader.StringKey(teamSlug.String()))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*db.Team), nil
}

func GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*db.UserRole, error) {
	loaders := For(ctx)
	thunk := loaders.UserRolesLoader.Load(ctx, dataloader.StringKey(userID.String()))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]*db.UserRole), nil
}