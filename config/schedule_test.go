package config

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduleMarshall(t *testing.T) {

	const jsonData = `{
		"id":0,
		"workflow_id":0,
		"channel_id":3038371147,
		"startDate":"2024-02-05T00:31:00-05:00",
		"frequency":1,
		"interval":0
	}`

	dataReader := strings.NewReader(jsonData)

	var schedule Schedule
	decoder := json.NewDecoder(dataReader)
	err := decoder.Decode(&schedule)
	assert.Nil(t, err)
}

func TestScheduleUnmarshall(t *testing.T) {

	schedule := NewSchedule()
	schedule.SetID(1)
	schedule.SetWorkflowID(2)
	schedule.SetChannelID(3)
	schedule.SetStartDate(time.Now())
	schedule.SetFrequency(1)
	schedule.SetInterval(2)

	jsonData, err := json.Marshal(schedule)
	assert.Nil(t, err)
	//assert.Equal(t, jsonData.GetID(), schedule.GetID())

	println(string(jsonData))
}
