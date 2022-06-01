package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryUsersOrderInput_GetOrderString(t *testing.T) {
	order := QueryUsersOrderInput{
		Field:     "name",
		Direction: "ASC",
	}
	assert.Equal(t, "name ASC", order.GetOrderString())
}
