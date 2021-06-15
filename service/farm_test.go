// +build broken

package service

import (
	"encoding/json"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/stretchr/testify/assert"
)

func TestFarmSetDeviceState(t *testing.T) {
	farm := mock(DefaultFarmService)

	farm.SetDeviceState("test", &state.DeviceState{
		Channels: []int{},
		Metrics:  map[int]float64{}})

	state, err := json.Marshal(deviceState)

	assert.Nil(t, err)
	assert.NotNil(t, state)
}
