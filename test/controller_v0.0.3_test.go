package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var deviceDataJSONv003 = `{
	"metrics": {
      "mem": 1200,
	  "sensor1": 12.34,
	  "sensor2": 56
	},
	"channels": [
		1, 1, 0, 1, 1
	]
}`

type TestDevicev003Entity struct {
	FreeMemory int     `json:"mem"`
	Sensor1    float64 `json:"sensor1"`
	Sensor2    float64 `json:"sensor2"`
}

type DeviceV003ConcreteState struct {
	Metrics  *TestDevicev003Entity `json:"metrics"`
	Channels []int                     `json:"channels"`
}

type DeviceV003DynamicState struct {
	Metrics  interface{} `json:"metrics"`
	Channels []int       `json:"channels"`
}

type DeviceV003State struct {
	Metrics  map[string]float64 `json:"metrics"`
	Channels []int              `json:"channels"`
}

func TestUnmarshallToState(t *testing.T) {

	var deviceState DeviceV003State
	err := json.Unmarshal([]byte(deviceDataJSONv003), &deviceState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	assert.Equal(t, 1200.0, deviceState.Metrics["mem"])
	assert.Equal(t, 12.34, deviceState.Metrics["sensor1"])
	assert.Equal(t, 56.0, deviceState.Metrics["sensor2"])

	assert.Equal(t, 1, deviceState.Channels[0])
	assert.Equal(t, 1, deviceState.Channels[1])
	assert.Equal(t, 0, deviceState.Channels[2])
	assert.Equal(t, 1, deviceState.Channels[3])
	assert.Equal(t, 1, deviceState.Channels[4])
}

func TestUnmarshallConcreteState(t *testing.T) {
	var deviceState DeviceV003ConcreteState
	err := json.Unmarshal([]byte(deviceDataJSONv003), &deviceState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	metrics := deviceState.Metrics

	assert.Equal(t, 1200, metrics.FreeMemory)
	assert.Equal(t, 12.34, metrics.Sensor1)
	assert.Equal(t, 56.0, metrics.Sensor2)

	assert.Equal(t, 5, len(deviceState.Channels))
	assert.Equal(t, 1, deviceState.Channels[0])
	assert.Equal(t, 1, deviceState.Channels[1])
	assert.Equal(t, 0, deviceState.Channels[2])
	assert.Equal(t, 1, deviceState.Channels[3])
	assert.Equal(t, 1, deviceState.Channels[4])
}

func TestUnmarshallDynamicStateWithConcreteEntity(t *testing.T) {
	var deviceState DeviceV003DynamicState
	err := json.Unmarshal([]byte(deviceDataJSONv003), &deviceState)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)

	tmp, err := json.Marshal(deviceState.Metrics)
	assert.Nil(t, err)

	metrics := &TestDevicev003Entity{}
	err = json.Unmarshal(tmp, metrics)
	assert.Nil(t, err)

	assert.Equal(t, 1200, metrics.FreeMemory)
	assert.Equal(t, 12.34, metrics.Sensor1)
	assert.Equal(t, 56.0, metrics.Sensor2)

	assert.Equal(t, 5, len(deviceState.Channels))
	assert.Equal(t, 1, deviceState.Channels[0])
	assert.Equal(t, 1, deviceState.Channels[1])
	assert.Equal(t, 0, deviceState.Channels[2])
	assert.Equal(t, 1, deviceState.Channels[3])
	assert.Equal(t, 1, deviceState.Channels[4])
}
