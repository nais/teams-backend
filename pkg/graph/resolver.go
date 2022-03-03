package graph

import (
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *gorm.DB
}

func NewResolver(db *gorm.DB) *Resolver {
	return &Resolver{
		db: db,
	}
}

// Limit a query by its pagination parameters, count number of rows in dataset, and return pagination metadata.
func (r *Resolver) withPagination(input model.PaginatedQuery, tx *gorm.DB) *model.Pagination {
	var count int64
	in := input.GetPagination()
	tx.Count(&count)
	tx.Limit(in.Limit)
	tx.Offset(in.Offset)
	return &model.Pagination{
		Results: int(count),
		Offset:  in.Offset,
		Limit:   in.Limit,
	}
}
