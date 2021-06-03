package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var controllerDataJSONv003 = `{
	"metrics": {
      "mem": 1200,
	  "sensor1": 12.34,
	  "sensor2": 56
	},
	"channels": [
		1, 1, 0, 1, 1
	]
}`

type TestControllerv003Entity struct {
	FreeMemory int     `json:"mem"`
	Sensor1    float64 `json:"sensor1"`
	Sensor2    float64 `json:"sensor2"`
}

type ControllerV003ConcreteState struct {
	Metrics  *TestControllerv003Entity `json:"metrics"`
	Channels []int                     `json:"channels"`
}

type ControllerV003DynamicState struct {
	Metrics  interface{} `json:"metrics"`
	Channels []int       `json:"channels"`
}

type ControllerV003State struct {
	Metrics  map[string]float64 `json:"metrics"`
	Channels []int              `json:"channels"`
}

func TestUnmarshallToState(t *testing.T) {

	var controllerState ControllerV003State
	err := json.Unmarshal([]byte(controllerDataJSONv003), &controllerState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	assert.Equal(t, 1200.0, controllerState.Metrics["mem"])
	assert.Equal(t, 12.34, controllerState.Metrics["sensor1"])
	assert.Equal(t, 56.0, controllerState.Metrics["sensor2"])

	assert.Equal(t, 1, controllerState.Channels[0])
	assert.Equal(t, 1, controllerState.Channels[1])
	assert.Equal(t, 0, controllerState.Channels[2])
	assert.Equal(t, 1, controllerState.Channels[3])
	assert.Equal(t, 1, controllerState.Channels[4])
}

func TestUnmarshallConcreteState(t *testing.T) {
	var controllerState ControllerV003ConcreteState
	err := json.Unmarshal([]byte(controllerDataJSONv003), &controllerState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	metrics := controllerState.Metrics

	assert.Equal(t, 1200, metrics.FreeMemory)
	assert.Equal(t, 12.34, metrics.Sensor1)
	assert.Equal(t, 56.0, metrics.Sensor2)

	assert.Equal(t, 5, len(controllerState.Channels))
	assert.Equal(t, 1, controllerState.Channels[0])
	assert.Equal(t, 1, controllerState.Channels[1])
	assert.Equal(t, 0, controllerState.Channels[2])
	assert.Equal(t, 1, controllerState.Channels[3])
	assert.Equal(t, 1, controllerState.Channels[4])
}

func TestUnmarshallDynamicStateWithConcreteEntity(t *testing.T) {
	var controllerState ControllerV003DynamicState
	err := json.Unmarshal([]byte(controllerDataJSONv003), &controllerState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	tmp, err := json.Marshal(controllerState.Metrics)
	assert.Nil(t, err)

	metrics := &TestControllerv003Entity{}
	err = json.Unmarshal(tmp, metrics)
	assert.Nil(t, err)

	assert.Equal(t, 1200, metrics.FreeMemory)
	assert.Equal(t, 12.34, metrics.Sensor1)
	assert.Equal(t, 56.0, metrics.Sensor2)

	assert.Equal(t, 5, len(controllerState.Channels))
	assert.Equal(t, 1, controllerState.Channels[0])
	assert.Equal(t, 1, controllerState.Channels[1])
	assert.Equal(t, 0, controllerState.Channels[2])
	assert.Equal(t, 1, controllerState.Channels[3])
	assert.Equal(t, 1, controllerState.Channels[4])
}
