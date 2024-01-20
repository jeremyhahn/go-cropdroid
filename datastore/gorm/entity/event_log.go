package entity

import "time"

type EventLogEntity interface {
	GetFarmID() uint64
	GetDevice() string
	GetType() string
	GetMessage() string
	GetTimestamp() string
	GetTimestampAsObject() time.Time
}

type EventLog struct {
	FarmID         uint64    `gorm:"not null" json:"farm_id"`
	Device         string    `gorm:"not null" json:"device"`
	Type           string    `gorm:"not null" json:"type"`
	Message        string    `gorm:"not null" json:"message"`
	Timestamp      time.Time `gorm:"type:timestamp" json:"timestamp"`
	EventLogEntity `json:"-"`
}

func (entity *EventLog) GetFarmID() uint64 {
	return entity.FarmID
}

func (entity *EventLog) GetDevice() string {
	return entity.Device
}

func (entity *EventLog) GetType() string {
	return entity.Type
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
