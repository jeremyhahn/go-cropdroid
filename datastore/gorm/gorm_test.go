package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGormConnection(t *testing.T) {
	currentTest := NewIntegrationTest()
	assert.NotNil(t, currentTest)
	currentTest.Cleanup()
}
