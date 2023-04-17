package dependencytrack

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDependencytrackReconciler_Reconcile(t *testing.T) {
	//database := db.NewMockDatabase(t)
	//auditLogger := auditlogger.NewMockAuditLogger(t)
	//log := logger.NewMockLogger(t)

	dptrackMock := newMock()

	cfg := &config.Config{
		DependencyTrack: config.DependencyTrack{
			Endpoint: "",
			Username: "",
			Password: "",
		},
	}

	reconciler, err := NewFromConfig(context.Background(), cfg)
	assert.NoError(t, err)

	for _, tt := range []struct {
		name   string
		team   string
		member string
	}{
		{
			name:   "team exists with a valid member",
			team:   "yolo",
			member: "nybruker@dev-nais.io",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			team := db.Team{
				Team: &sqlc.Team{
					Slug:    slug.Slug(tt.team),
					Purpose: "teamPurpose",
				},
			}

			members := []*db.User{
				{User: &sqlc.User{
					Email: tt.member,
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
		})
	}
}

func teamExists(mock *mock, team string) bool {
	for _, t := range mock.teams {
		if t.Name == team {
			return true
		}
	}

	return false
}
