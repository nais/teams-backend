package dbmodels

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

type stateContainer struct {
	Value string `json:"value"`
}

func newUuid() uuid.UUID {
	id, _ := uuid.NewUUID()
	return id
}

func TestLoadSystemState(t *testing.T) {
	systemId := newUuid()
	teamId := newUuid()

	t.Run("No existing state", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(SystemState{})

		state := &stateContainer{}
		assert.NoError(t, LoadSystemState(db, systemId, teamId, state))
		assert.Equal(t, "", state.Value)
		assert.NoError(t, SetSystemState(db, systemId, teamId, stateContainer{Value: "some value"}))
		assert.NoError(t, LoadSystemState(db, systemId, teamId, state))
		assert.Equal(t, "some value", state.Value)
	})

	t.Run("Direct update", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(SystemState{})

		state := &stateContainer{}
		assert.NoError(t, SetSystemState(db, systemId, teamId, stateContainer{Value: "some value"}))
		assert.NoError(t, LoadSystemState(db, systemId, teamId, state))
		assert.Equal(t, "some value", state.Value)
	})
}
