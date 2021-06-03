// +build broken

package service

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestFarmSetControllerState(t *testing.T) {
	farm := mock(DefaultFarmService)

	farm.SetControllerState("test", &state.ControllerState{
		Channels: []int{},
		Metrics:  map[int]float64{}})

	state, err := json.Marshal(controllerState)

	assert.Nil(t, err)
	assert.NotNil(t, state)
}
