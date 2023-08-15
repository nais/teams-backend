package helpers_test

import (
	"testing"

	"github.com/nais/teams-backend/pkg/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStringWithFallback(t *testing.T) {
	t.Run("Fallback not used", func(t *testing.T) {
		assert.Equal(t, "some value", helpers.StringWithFallback(helpers.Strp("some value"), "some fallback value"))
	})

	t.Run("Fallback used", func(t *testing.T) {
		assert.Equal(t, "some fallback value", helpers.StringWithFallback(helpers.Strp(""), "some fallback value"))
	})
}

func TestStrp(t *testing.T) {
	s := "some string"
	assert.Equal(t, &s, helpers.Strp(s))
}

func TestBoolp(t *testing.T) {
	b := true
	assert.Equal(t, &b, helpers.Boolp(b))
}

func TestTruncate(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		assert.Equal(t, "", helpers.Truncate("", 5))
	})

	t.Run("String shorter than truncate length", func(t *testing.T) {
		assert.Equal(t, "some string", helpers.Truncate("some string", 20))
	})

	t.Run("String longer than truncate length", func(t *testing.T) {
		assert.Equal(t, "some ", helpers.Truncate("some string", 5))
	})
}

func TestContains(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		assert.False(t, helpers.Contains([]string{}, "string"))
	})

	t.Run("Slice contains string", func(t *testing.T) {
		assert.True(t, helpers.Contains([]string{"foo", "bar"}, "bar"))
	})

	t.Run("Slice does not contain string", func(t *testing.T) {
		assert.False(t, helpers.Contains([]string{"foo", "bar"}, "Bar"))
	})
}
