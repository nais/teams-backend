package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSortInputs(t *testing.T) {
	t.Run("Users sorting", func(t *testing.T) {
		order := QueryUsersSortInput{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})

	t.Run("Teams sorting", func(t *testing.T) {
		order := QueryTeamsSortInput{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})
}
