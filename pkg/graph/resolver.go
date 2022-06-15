package graph

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const (
	DuplicateErrorCode = "23505"
)

type Resolver struct {
	db             *gorm.DB
	partnerDomain  string
	teamReconciler chan<- reconcilers.ReconcileTeamInput
	system         *dbmodels.System
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(db *gorm.DB, partnerDomain string, system *dbmodels.System, teamReconciler chan<- reconcilers.ReconcileTeamInput, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		db:             db,
		partnerDomain:  partnerDomain,
		system:         system,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}

// Model Enables abstracted access to CreatedBy and UpdatedBy for generic database models.
type Model interface {
	GetModel() *dbmodels.Model
}

func (r *Resolver) createTrackedObject(ctx context.Context, obj Model) error {
	user := authz.UserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("context has no user")
	}
	model := obj.GetModel()
	model.CreatedBy = user
	model.UpdatedBy = user
	return r.db.WithContext(ctx).Create(obj).Error
}

func (r *Resolver) createTrackedObjectIgnoringDuplicates(ctx context.Context, obj Model) error {
	err := r.createTrackedObject(ctx, obj)

	if err == nil {
		return nil
	}

	switch t := err.(type) {
	case *pgconn.PgError:
		if t.Code == DuplicateErrorCode {
			return nil
		}
	}

	return err
}

// Update an object in the database, attaching the current user in metadata.
func (r *Resolver) updateDB(ctx context.Context, obj Model) error {
	m := obj.GetModel()
	m.UpdatedBy = authz.UserFromContext(ctx)
	tx := r.db.WithContext(ctx).Updates(obj)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no such %T", obj)
	}
	return tx.Find(obj).Error
}

// Run a query to get data from the database. Populates `collection` and returns pagination metadata.
func (r *Resolver) paginatedQuery(ctx context.Context, pagination *model.Pagination, query model.Query, sort model.QueryOrder, dbmodel interface{}, collection interface{}) (*model.PageInfo, error) {
	if pagination == nil {
		pagination = &model.Pagination{
			Offset: 0,
			Limit:  50,
		}
	}
	tx := r.db.WithContext(ctx).Model(dbmodel).Where(query.GetQuery()).Order(sort.GetOrderString())
	pageInfo, tx := r.withPagination(pagination, tx)
	return pageInfo, tx.Find(collection).Error
}

// Limit a query by its pagination parameters, count number of rows in dataset, and return pagination metadata.
func (r *Resolver) withPagination(pagination *model.Pagination, tx *gorm.DB) (*model.PageInfo, *gorm.DB) {
	var count int64
	tx = tx.Count(&count).Limit(pagination.Limit).Offset(pagination.Offset)

	return &model.PageInfo{
		Results: int(count),
		Offset:  pagination.Offset,
		Limit:   pagination.Limit,
	}, tx
}

func (r *mutationResolver) teamWithAssociations(ctx context.Context, teamID uuid.UUID) (*dbmodels.Team, error) {
	team := &dbmodels.Team{}
	err := r.db.
		WithContext(ctx).
		Where("id = ?", teamID).
		Preload("Users").
		Preload("SystemState").
		Preload("Metadata").
		First(team).Error

	if err != nil {
		return nil, err
	}

	return team, nil
}
