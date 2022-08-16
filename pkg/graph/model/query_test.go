package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSortInputs(t *testing.T) {
	t.Run("Users sorting", func(t *testing.T) {
		order := UsersSort{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})

	t.Run("Teams sorting", func(t *testing.T) {
		order := TeamsSort{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})

	t.Run("Audit logs sorting", func(t *testing.T) {
		order := AuditLogsSort{
			Field:     "name",
			Direction: "ASC",
		}
		assert.Equal(t, "name ASC", order.GetOrderString())
	})
}
