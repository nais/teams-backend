package dependencytrack

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"testing"
)

func TestDependencytrackReconciler_Reconcile(t *testing.T) {
	//database := db.NewMockDatabase(t)
	//auditLogger := auditlogger.NewMockAuditLogger(t)
	//log := logger.NewMockLogger(t)
	reconciler, err := NewFromConfig(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	team := db.Team{
		Team: &sqlc.Team{
			Slug:    slug.Slug("yolo"),
			Purpose: "teamPurpose",
		},
	}
	members := []*db.User{
		{User: &sqlc.User{
			Email: "nybruker@dev-nais.io",
		},
		},
	}
	input := reconcilers.Input{
		CorrelationID:   uuid.New(),
		Team:            team,
		TeamMembers:     members,
		NumSyncAttempts: 0,
	}

	err = reconciler.Reconcile(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
}
