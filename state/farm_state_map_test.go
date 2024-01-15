package state

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFarmStateMap(t *testing.T) {

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	deviceStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	farmStateMap := NewFarmStateMap(1)
	farmStateMap.SetDevice("testdevice", deviceStateMap)

	state, err := farmStateMap.GetDevice("testdevice")
	assert.Nil(t, err)
	assert.Equal(t, deviceStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])
}

func TestSerialization(t *testing.T) {

	deviceStateMap := NewDeviceStateMap()
	deviceStateMap.SetMetrics(map[string]float64{
		"test":  12.34,
		"test2": 56.7})
	deviceStateMap.SetChannels([]int{1, 0, 1, 0, 1, 1})

	t2, _ := json.Marshal(deviceStateMap)
	println(string(t2))

	farmStateMap := NewFarmStateMap(1)
	farmStateMap.SetDevice("testdevice", deviceStateMap)

	println(fmt.Sprintf("%+v", farmStateMap))
	println(fmt.Sprintf("%+v", farmStateMap.GetDevices()))
	println(fmt.Sprintf("%+v", farmStateMap.GetDevices()["testdevice"]))

	state, err := farmStateMap.GetDevice("testdevice")
	assert.Nil(t, err)
	assert.Equal(t, deviceStateMap, state)
	assert.Equal(t, 12.34, state.GetMetrics()["test"])
	assert.Equal(t, 56.7, state.GetMetrics()["test2"])
	assert.Equal(t, 1, state.GetChannels()[0])
	assert.Equal(t, 0, state.GetChannels()[1])

	data, err := json.Marshal(farmStateMap)
	assert.Nil(t, err)
	assert.NotNil(t, data)
	println(string(data))
}

func TestUnserialize(t *testing.T) {
	data := `{"id":562173130332602369,"devices":{"room":{"metrics":{"co2":2318.75,"heatIndex0":84.4,"heatIndex1":24.9,"heatIndex2":24.9,"humidity0":60,"humidity1":0,"humidity2":0,"leak0":2,"leak1":0,"mem":1500,"photo":0,"tempC0":24.3,"tempC1":0,"tempC2":0,"tempF0":75,"tempF1":32,"tempF2":32,"vpd":-0.51,"water0":65.19,"water1":-196.6},"channels":[1,0,0,1,0,1,0,0,0],"timestamp":"2020-06-14T17:19:27.103050033-04:00"}},"timestamp":1592169566}`
	var fs FarmState
	err := json.Unmarshal([]byte(data), &fs)
	assert.Nil(t, err)
	assert.NotNil(t, fs)
	assert.NotNil(t, fs.Devices)
	assert.Equal(t, uint64(562173130332602369), fs.ID)
	assert.Equal(t, 2318.75, fs.Devices["room"].GetMetrics()["co2"])
	assert.Equal(t, float64(60), fs.Devices["room"].GetMetrics()["humidity0"])
	assert.Equal(t, 1, fs.Devices["room"].GetChannels()[0])
	assert.Equal(t, 0, fs.Devices["room"].GetChannels()[1])
}
