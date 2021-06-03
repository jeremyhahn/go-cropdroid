package test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var controllerDataJSON = `{
    "mem": 1200,
	"sensor1": 12.34,
	"sensor2": 56,
	"channels": {
		"0": 1,
		"1": 1,
		"2": 1,
		"3": 1,
		"4": 1
	}
}`

type TestChannels struct {
	Channel0 int `json:"0"`
	Channel1 int `json:"1"`
	Channel2 int `json:"2"`
	Channel3 int `json:"3"`
	Channel4 int `json:"4"`
}

type TestControllerDataEntity struct {
	FreeMemory int           `gorm:"column:mem" json:"mem"`
	Sensor1    float64       `gorm:"column:sensor1" json:"sensor1"`
	Sensor2    float64       `gorm:"column:sensor2" json:"sensor2"`
	Channels   *TestChannels `gorm:"-" json:"channels"`
}

func createTestInterfaceEntity() interface{} {
	return &TestControllerDataEntity{}
}

func TestUnmarshallEntity(t *testing.T) {
	var entity TestControllerDataEntity
	err := json.Unmarshal([]byte(controllerDataJSON), &entity)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	assert.Nil(t, err)
	assert.Equal(t, entity.FreeMemory, 1200)
	assert.Equal(t, entity.Sensor1, 12.34)
	assert.Equal(t, 1, entity.Channels.Channel0)
	assert.Equal(t, 1, entity.Channels.Channel1)
	assert.Equal(t, 1, entity.Channels.Channel2)
	assert.Equal(t, 1, entity.Channels.Channel3)
	assert.Equal(t, 1, entity.Channels.Channel4)
}

func TestDynamicEntityWithInterface(t *testing.T) {
	data := createTestInterfaceEntity()
	err := json.Unmarshal([]byte(controllerDataJSON), data)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	var entity = data.(*TestControllerDataEntity)
	assert.Nil(t, err)
	assert.Equal(t, entity.FreeMemory, 1200)
	assert.Equal(t, entity.Sensor1, 12.34)
	assert.Equal(t, 1, entity.Channels.Channel0)
	assert.Equal(t, 1, entity.Channels.Channel1)
	assert.Equal(t, 1, entity.Channels.Channel2)
	assert.Equal(t, 1, entity.Channels.Channel3)
	assert.Equal(t, 1, entity.Channels.Channel4)
}

func TestDynamicEntityWithOriginalControllerAPI(t *testing.T) {

	metricMap := make(map[string]float64, 3)
	channelMap := make(map[int]int, 5)

	data := createTestInterfaceEntity()
	err := json.Unmarshal([]byte(controllerDataJSON), data)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	var dataEntity = data.(*TestControllerDataEntity)

	reflector := reflect.ValueOf(*dataEntity)
	typeOf := reflector.Type()
	for i := 0; i < reflector.NumField(); i++ {

		field := typeOf.Field(i).Name
		value := reflector.Field(i).Interface()
		jsonField := typeOf.Field(i).Tag.Get("json")
		fmt.Printf("Field: %s\tValue: %v\tJSON Field:%s\n", field, value, jsonField)

		if jsonField == "channels" {
			channelReflector := reflect.ValueOf(*value.(*TestChannels))
			channelTypeOf := channelReflector.Type()
			for j := 0; j < channelReflector.NumField(); j++ {
				channelField := channelTypeOf.Field(j).Name
				channelValue := channelReflector.Field(j).Interface()
				channelJsonField := channelTypeOf.Field(j).Tag.Get("json")
				fmt.Printf("\tChannel Field: %s\tValue: %v\tJSON Field:%s\n", channelField, channelValue, channelJsonField)

				channelMap[j] = channelValue.(int)
				continue
			}
		}

		var v float64
		switch reflector.Field(i).Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			v = float64(value.(int))
		case reflect.Float32, reflect.Float64:
			v = value.(float64)
		default:
			fmt.Printf("Unsupported metric data type: %s\n", field)
			continue
			//assert.Fail(t, "Unsupported metric data type: "+field)
		}

		metricMap[jsonField] = v
	}

	//fmt.Println(metricMap)
	//fmt.Println(channelMap)

	assert.Nil(t, err)
	assert.Equal(t, dataEntity.FreeMemory, 1200)
	assert.Equal(t, dataEntity.Sensor1, 12.34)
	assert.Equal(t, 1, dataEntity.Channels.Channel0)
	assert.Equal(t, 1, dataEntity.Channels.Channel1)
	assert.Equal(t, 1, dataEntity.Channels.Channel2)
	assert.Equal(t, 1, dataEntity.Channels.Channel3)
	assert.Equal(t, 1, dataEntity.Channels.Channel4)

	assert.Equal(t, float64(1200), metricMap["mem"])
	assert.Equal(t, float64(12.34), metricMap["sensor1"])
	assert.Equal(t, float64(56), metricMap["sensor2"])
	assert.Equal(t, float64(0), metricMap["channels"])

	assert.Equal(t, 1, channelMap[0])
	assert.Equal(t, 1, channelMap[1])
	assert.Equal(t, 1, channelMap[2])
	assert.Equal(t, 1, channelMap[3])
	assert.Equal(t, 1, channelMap[4])
}
