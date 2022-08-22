package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
)

func (r *aPIKeyResolver) APIKey(ctx context.Context, obj *db.APIKey) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateAPIKey(ctx context.Context, userID *uuid.UUID) (*db.APIKey, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteAPIKey(ctx context.Context, userID *uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

// APIKey returns generated.APIKeyResolver implementation.
func (r *Resolver) APIKey() generated.APIKeyResolver { return &aPIKeyResolver{r} }

type aPIKeyResolver struct{ *Resolver }
