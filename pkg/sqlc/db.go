// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.createCorrelationStmt, err = db.PrepareContext(ctx, createCorrelation); err != nil {
		return nil, fmt.Errorf("error preparing query CreateCorrelation: %w", err)
	}
	if q.createSystemStmt, err = db.PrepareContext(ctx, createSystem); err != nil {
		return nil, fmt.Errorf("error preparing query CreateSystem: %w", err)
	}
	if q.getRoleStmt, err = db.PrepareContext(ctx, getRole); err != nil {
		return nil, fmt.Errorf("error preparing query GetRole: %w", err)
	}
	if q.getRolesStmt, err = db.PrepareContext(ctx, getRoles); err != nil {
		return nil, fmt.Errorf("error preparing query GetRoles: %w", err)
	}
	if q.getSystemStmt, err = db.PrepareContext(ctx, getSystem); err != nil {
		return nil, fmt.Errorf("error preparing query GetSystem: %w", err)
	}
	if q.getSystemByNameStmt, err = db.PrepareContext(ctx, getSystemByName); err != nil {
		return nil, fmt.Errorf("error preparing query GetSystemByName: %w", err)
	}
	if q.getSystemsStmt, err = db.PrepareContext(ctx, getSystems); err != nil {
		return nil, fmt.Errorf("error preparing query GetSystems: %w", err)
	}
	if q.getUserStmt, err = db.PrepareContext(ctx, getUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetUser: %w", err)
	}
	if q.getUserRoleStmt, err = db.PrepareContext(ctx, getUserRole); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserRole: %w", err)
	}
	if q.getUserRolesStmt, err = db.PrepareContext(ctx, getUserRoles); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserRoles: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.createCorrelationStmt != nil {
		if cerr := q.createCorrelationStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createCorrelationStmt: %w", cerr)
		}
	}
	if q.createSystemStmt != nil {
		if cerr := q.createSystemStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createSystemStmt: %w", cerr)
		}
	}
	if q.getRoleStmt != nil {
		if cerr := q.getRoleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getRoleStmt: %w", cerr)
		}
	}
	if q.getRolesStmt != nil {
		if cerr := q.getRolesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getRolesStmt: %w", cerr)
		}
	}
	if q.getSystemStmt != nil {
		if cerr := q.getSystemStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getSystemStmt: %w", cerr)
		}
	}
	if q.getSystemByNameStmt != nil {
		if cerr := q.getSystemByNameStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getSystemByNameStmt: %w", cerr)
		}
	}
	if q.getSystemsStmt != nil {
		if cerr := q.getSystemsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getSystemsStmt: %w", cerr)
		}
	}
	if q.getUserStmt != nil {
		if cerr := q.getUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserStmt: %w", cerr)
		}
	}
	if q.getUserRoleStmt != nil {
		if cerr := q.getUserRoleStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserRoleStmt: %w", cerr)
		}
	}
	if q.getUserRolesStmt != nil {
		if cerr := q.getUserRolesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserRolesStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                    DBTX
	tx                    *sql.Tx
	createCorrelationStmt *sql.Stmt
	createSystemStmt      *sql.Stmt
	getRoleStmt           *sql.Stmt
	getRolesStmt          *sql.Stmt
	getSystemStmt         *sql.Stmt
	getSystemByNameStmt   *sql.Stmt
	getSystemsStmt        *sql.Stmt
	getUserStmt           *sql.Stmt
	getUserRoleStmt       *sql.Stmt
	getUserRolesStmt      *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                    tx,
		tx:                    tx,
		createCorrelationStmt: q.createCorrelationStmt,
		createSystemStmt:      q.createSystemStmt,
		getRoleStmt:           q.getRoleStmt,
		getRolesStmt:          q.getRolesStmt,
		getSystemStmt:         q.getSystemStmt,
		getSystemByNameStmt:   q.getSystemByNameStmt,
		getSystemsStmt:        q.getSystemsStmt,
		getUserStmt:           q.getUserStmt,
		getUserRoleStmt:       q.getUserRoleStmt,
		getUserRolesStmt:      q.getUserRolesStmt,
	}
}
