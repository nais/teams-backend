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

	t.Run("Audit logs sorting", func(t *testing.T) {
		order := QueryAuditLogsSortInput{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})

	t.Run("System sorting", func(t *testing.T) {
		order := QuerySystemsSortInput{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})
}
