package model

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type Notification struct {
	Device              string    `json:"device"`
	Priority            int       `json:"priority"`
	Type                string    `json:"type"`
	Title               string    `json:"title"`
	Message             string    `json:"message"`
	Timestamp           time.Time `json:"timestamp"`
	common.Notification `json:"-"`
}

func (model *Notification) GetDevice() string {
	return model.Device
}

func (model *Notification) GetPriority() int {
	return model.Priority
}

func (model *Notification) GetType() string {
	return model.Type
}

func (model *Notification) GetTitle() string {
	return model.Title
}

func (model *Notification) GetMessage() string {
	return model.Message
}

func (model *Notification) GetTimestamp() string {
	return model.Timestamp.Format(time.RFC3339)
}

func (model *Notification) GetTimestampAsObject() time.Time {
	return model.Timestamp
}
