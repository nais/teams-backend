package test

import (
	"testing"
)

func TestGetTestDBCanRunMultipleTimes(t *testing.T) {
	_ = GetTestDB()
	_ = GetTestDB()
}
