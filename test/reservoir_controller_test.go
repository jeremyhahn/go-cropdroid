// +build integration

package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/controller"
	"github.com/stretchr/testify/assert"
)

func TestReservoirClient(t *testing.T) {

	ctx := NewUnitTestContext()

	client := controller.NewReservoirController(ctx, "http://res1.westland.dr", nil)
	entity, err := client.ReservoirStatus()

	assert.Nil(t, err)
	assert.True(t, entity.GetPH() > 0)
}
