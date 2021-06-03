package model

import (
	"time"

	"github.com/jeremyhahn/cropdroid/common"
)

type Notification struct {
	Controller          string    `json:"controller"`
	Priority            int       `json:"priority"`
	Type                string    `json:"type"`
	Title               string    `json:"title"`
	Message             string    `json:"message"`
	Timestamp           time.Time `json:"timestamp"`
	common.Notification `json:"-"`
}

func (model *Notification) GetController() string {
	return model.Controller
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
