package state

import (
	"testing"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestControllerIndex(t *testing.T) {

	controllerConfig := &config.Controller{
		ID:          1,
		FarmID:      1,
		Type:        "Test",
		Interval:    60,
		Description: "This is a controller used for testing",
		Enable:      true,
		Notify:      false,
		URI:         "http://localhost"}

	store := NewControllerIndex(1)
	store.Put(1, controllerConfig)

	controller, ok := store.Get(1)
	assert.Equal(t, true, ok)
	assert.Equal(t, controller, controllerConfig)
	assert.Equal(t, 1, controller.GetID())
	assert.Equal(t, 1, controller.GetFarmID())
	assert.Equal(t, "Test", controller.GetType())
}

func BenchmarkControllerIndex(b *testing.B) {

	controllerConfig := &config.Controller{
		ID:          1,
		FarmID:      1,
		Type:        "Test",
		Interval:    60,
		Description: "This is a controller used for testing",
		Enable:      true,
		Notify:      false,
		URI:         "http://localhost"}

	store := NewControllerIndex(b.N)

	for n := 0; n < b.N; n++ {
		store.Put(n, controllerConfig)
	}
}
