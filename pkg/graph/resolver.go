package graph

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/roles"
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
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	system         *dbmodels.System
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(db *gorm.DB, tenantDomain string, system *dbmodels.System, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		db:             db,
		tenantDomain:   tenantDomain,
		system:         system,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}

// Model Enables abstracted access to CreatedBy and UpdatedBy for generic database models.
type Model interface {
	GetModel() *dbmodels.Model
}

func (r *Resolver) createTrackedObject(ctx context.Context, newObject Model) error {
	user := authz.UserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("context has no user")
	}
	model := newObject.GetModel()
	model.CreatedBy = user
	model.UpdatedBy = user
	return r.db.Create(newObject).Error
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

func (r *Resolver) updateTrackedObject(ctx context.Context, updatedObject Model) error {
	user := authz.UserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("context has no user")
	}

	model := updatedObject.GetModel()
	model.UpdatedBy = user
	return r.db.Updates(updatedObject).Error
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

func (r *Resolver) userIsTeamOwner(userId, teamId uuid.UUID) (bool, error) {
	teamOwnerRole := &dbmodels.Role{}
	err := r.db.Where("name = ?", roles.RoleTeamOwner).First(teamOwnerRole).Error
	if err != nil {
		return false, err
	}

	roleBinding := &dbmodels.UserRole{}
	err = r.db.Where("role_id = ? AND user_id = ? AND target_id = ?", teamOwnerRole.ID, userId, teamId).First(roleBinding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
