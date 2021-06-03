package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericStore(t *testing.T) {

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	controllerStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	store := NewGenericStore(1)
	store.Put(1, controllerStateMap)

	_state, ok := store.Get(1)
	assert.Equal(t, true, ok)

	state := _state.(ControllerStateMap)
	assert.Equal(t, controllerStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])
}

func BenchmarkGenericStorePut(b *testing.B) {

	controllerStateMap := NewControllerStateMap()
	controllerStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	controllerStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	store := NewGenericStore(b.N)

	for n := 0; n < b.N; n++ {
		store.Put(n, controllerStateMap)
	}
}
