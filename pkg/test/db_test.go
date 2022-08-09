package test

import (
	"testing"
)

func TestGetTestDBCanRunMultipleTimes(t *testing.T) {
	_, _ = GetTestDB()
	_, _ = GetTestDB()
}
