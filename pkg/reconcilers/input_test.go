package reconcilers_test

import (
	"testing"

	"github.com/nais/console/pkg/db"

	"github.com/nais/console/pkg/reconcilers"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestInputWithCorrelationID(t *testing.T) {
	id1 := uuid.New()
	input1 := &reconcilers.Input{CorrelationID: id1, Team: db.Team{}}

	id2 := uuid.New()
	input2 := input1.WithCorrelationID(id2)

	assert.Equal(t, id1, input1.CorrelationID)
	assert.Equal(t, id2, input2.CorrelationID)
	assert.Equal(t, input1.Team, input2.Team)
}
