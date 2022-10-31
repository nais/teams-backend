package db_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestRole_IsGlobal(t *testing.T) {
	targetID := uuid.New()
	r1 := db.Role{TargetID: &targetID}
	r2 := db.Role{}
	assert.False(t, r1.IsGlobal())
	assert.True(t, r2.IsGlobal())
}

func TestRole_Targets(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	r1 := db.Role{TargetID: &id1}
	r2 := db.Role{TargetID: &id2}
	r3 := db.Role{}

	assert.True(t, r1.Targets(id1))
	assert.False(t, r2.Targets(id1))
	assert.False(t, r3.Targets(id1))

	assert.False(t, r1.Targets(id2))
	assert.True(t, r2.Targets(id2))
	assert.False(t, r3.Targets(id2))
}
