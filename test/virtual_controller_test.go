// +build broken

package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/controller"
	"github.com/stretchr/testify/assert"
)

func TestVirtualRoomMapStateToEntity(t *testing.T) {

	app, _ := NewUnitTestContext()
	farmState := common.NewFarmStateMap()
	controllerType := "test"

	vcontroller := controller.NewVirtualController(app, farmState, "", controllerType)

	metrics := map[string]float64{
		"mem":        1200,
		"tempF0":     75.52,
		"tempC0":     24.17,
		"humidity0":  65.0,
		"heatIndex0": 63.2,
		"tempF1":     76.17,
		"tempC1":     25.17,
		"humidity1":  66.0,
		"heatIndex1": 64.3,
		"tempF2":     77.17,
		"tempC2":     26.17,
		"humidity2":  67.0,
		"heatIndex2": 65.0,
		"vpd":        .12,
		"water0":     64.0,
		"water1":     64.12,
		"co2":        1200.0,
		"leak0":      1,
		"leak1":      2,
		"photo":      800}

	channels := []int{0, 0, 1, 0, 0, 1}

	fakeRoomState := common.CreateControllerStateMap(metrics, channels)

	vcontroller.WriteState(fakeRoomState)

	assert.Equal(t, "test", vcontroller.GetType())
}
