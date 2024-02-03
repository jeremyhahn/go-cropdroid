package entity

import "time"

type EventLogEntity interface {
	GetFarmID() uint64
	GetDeviceID() uint64
	GetDeviceName() string
	GetEventType() string
	GetMessage() string
	GetTimestamp() string
	GetTimestampAsObject() time.Time
}

type EventLog struct {
	FarmID         uint64    `gorm:"not null" json:"farm_id"`
	DeviceID       uint64    `gorm:"not null" json:"device_id"`
	DeviceName     string    `gorm:"not null" json:"device"`
	EventType      string    `gorm:"not null" json:"type"`
	Message        string    `gorm:"not null" json:"message"`
	Timestamp      time.Time `gorm:"type:timestamp" json:"timestamp"`
	EventLogEntity `json:"-"`
}

func (entity *EventLog) GetFarmID() uint64 {
	return entity.FarmID
}

func (entity *EventLog) GetDeviceID() uint64 {
	return entity.DeviceID
}

func (entity *EventLog) GetNameDevice() string {
	return entity.DeviceName
}

func (entity *EventLog) GetEventType() string {
	return entity.EventType
}

func (entity *EventLog) GetMessage() string {
	return entity.Message
}

func (entity *EventLog) GetTimestamp() string {
	return entity.Timestamp.Format(time.RFC3339)
}

func (entity *EventLog) GetTimestampAsObject() time.Time {
	return entity.Timestamp
}
