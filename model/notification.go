package model

import (
	"time"
)

type Notification interface {
	GetDevice() string
	GetPriority() int
	GetType() string
	GetTitle() string
	GetMessage() string
	GetTimestamp() string
	GetTimestampAsObject() time.Time
}

type NotificationStruct struct {
	Device       string    `json:"device"`
	Priority     int       `json:"priority"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	Notification `json:"-"`
}

func (model *NotificationStruct) GetDevice() string {
	return model.Device
}

func (model *NotificationStruct) GetPriority() int {
	return model.Priority
}

func (model *NotificationStruct) GetType() string {
	return model.Type
}

func (model *NotificationStruct) GetTitle() string {
	return model.Title
}

func (model *NotificationStruct) GetMessage() string {
	return model.Message
}

func (model *NotificationStruct) GetTimestamp() string {
	return model.Timestamp.Format(time.RFC3339)
}

func (model *NotificationStruct) GetTimestampAsObject() time.Time {
	return model.Timestamp
}
