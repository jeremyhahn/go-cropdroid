// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/controller"
	"github.com/jeremyhahn/cropdroid/test"
	"github.com/stretchr/testify/assert"
)

var testReservoirController = "http://192.168.0.92"
var testReservoirControllerType = "reservoir"
var testReserviorMetricCount = 15
var testReserviorChannelCount = 16

func TestState(t *testing.T) {

	ctx := test.NewUnitTestContext()

	httpController := controller.NewHttpController(ctx, testReservoirController, testReservoirControllerType)

	state, err := httpController.State()
	assert.Nil(t, err)

	/*
		for k, v := range state.Metrics {
			fmt.Printf("metric key=%s, value=%.2f\n", k, v)
		}
		for i, channel := range state.Channels {
			fmt.Printf("channel id=%d, value=%d\n", i, channel)
		}*/

	assert.Equal(t, testReserviorMetricCount, len(state.Metrics))
	assert.Equal(t, testReserviorChannelCount, len(state.Channels))
}

func TestChannelSwitch(t *testing.T) {

	ctx := test.NewUnitTestContext()

	reservoirController := controller.NewHttpController(ctx, testReservoirController, testReservoirControllerType)

	initialState, err := reservoirController.State()
	assert.Nil(t, err)

	// Ensure switch state is OFF and turn it on
	for channelID, channelState := range initialState.Channels {
		assert.Equal(t, common.SWITCH_OFF, channelState)
		reservoirController.Switch(channelID, common.SWITCH_ON)
		//fmt.Printf("channel id=%d, state=%d\n", channelID, channelState)
	}

	// Ensure switch state is on
	newState, err := reservoirController.State()
	assert.Nil(t, err)
	for channelID, channelState := range newState.Channels {
		assert.Equal(t, common.SWITCH_ON, channelState)
		reservoirController.Switch(channelID, common.SWITCH_OFF)
	}

	// Cleanup by switching all channels back to OFF
	finalState, err := reservoirController.State()
	assert.Nil(t, err)
	for _, channelState := range finalState.Channels {
		assert.Equal(t, common.SWITCH_OFF, channelState)
	}
}

func TestChannelTimedSwitch(t *testing.T) {

	ctx := test.NewUnitTestContext()

	reservoirController := controller.NewHttpController(ctx, testReservoirController, testReservoirControllerType)

	initialState, err := reservoirController.State()
	assert.Nil(t, err)

	channelID := 0
	timer := 2 // seconds

	// Ensure initial state is OFF
	channelState := initialState.Channels[channelID]
	assert.Equal(t, common.SWITCH_OFF, channelState)

	// Switch on for 2 seconds
	reservoirController.TimedSwitch(channelID, timer)

	// Ensure switch is on
	newState, err := reservoirController.State()
	assert.Nil(t, err)
	assert.Equal(t, common.SWITCH_ON, newState.Channels[channelID])

	// Wait for the timer to expire
	time.Sleep(time.Duration(timer) * time.Second)

	// Ensure the switch state is OFF
	finalChannelState, err := reservoirController.State()
	assert.Equal(t, common.SWITCH_OFF, finalChannelState.Channels[channelID])
}
