package graph

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	queries        *sqlc.Queries
	db             *gorm.DB
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	system         sqlc.System
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(queries *sqlc.Queries, db *gorm.DB, tenantDomain string, system sqlc.System, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		queries:        queries,
		db:             db,
		tenantDomain:   tenantDomain,
		system:         system,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}

// createCorrelation Create a correlation entry in the database
func (r *Resolver) createCorrelation(ctx context.Context) (*sqlc.Correlation, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to generate ID for correlation")
	}
	correlation, err := r.queries.CreateCorrelation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to create correlation entry")
	}
	return correlation, nil
}

// Run a query to get data from the database. Populates `collection` and returns pagination metadata.
func (r *Resolver) paginatedQuery(pagination *model.Pagination, query model.Query, sort model.QueryOrder, dbModel interface{}, collection interface{}) (*model.PageInfo, error) {
	if pagination == nil {
		pagination = &model.Pagination{
			Offset: 0,
			Limit:  50,
		}
	}
	db := r.db.Model(dbModel).Where(query.GetQuery()).Order(sort.GetOrderString())
	pageInfo, db := r.withPagination(pagination, db)
	return pageInfo, db.Find(collection).Error
}

// Limit a query by its pagination parameters, count number of rows in dataset, and return pagination metadata.
func (r *Resolver) withPagination(pagination *model.Pagination, db *gorm.DB) (*model.PageInfo, *gorm.DB) {
	var count int64
	db = db.Count(&count).Limit(pagination.Limit).Offset(pagination.Offset)

	return &model.PageInfo{
		Results: int(count),
		Offset:  pagination.Offset,
		Limit:   pagination.Limit,
	}, db
}

func (r *mutationResolver) teamWithAssociations(teamID uuid.UUID) (*dbmodels.Team, error) {
	team := &dbmodels.Team{}
	err := r.db.
		Where("id = ?", teamID).
		Preload("Users").
		Preload("Metadata").
		First(team).Error

	if err != nil {
		return nil, err
	}

	return team, nil
}
