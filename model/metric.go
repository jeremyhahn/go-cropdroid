package model

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type Metric struct {
	ID            uint64     `yaml:"id" json:"id"`
	DeviceID      uint64     `yaml:"deviceID" json:"deviceId"`
	DataType      int        `yaml:"datatype" json:"datatype"`
	Name          string     `yaml:"name" json:"name"`
	Key           string     `yaml:"key" json:"key"`
	Enable        bool       `yaml:"enable" json:"enable"`
	Notify        bool       `yaml:"notify" json:"notify"`
	Unit          string     `yaml:"unit" json:"unit"`
	AlarmLow      float64    `yaml:"alarmLow" json:"alarmLow"`
	AlarmHigh     float64    `yaml:"alarmHigh" json:"alarmHigh"`
	Value         float64    `json:"value"`
	Timestamp     *time.Time `json:"-"`
	common.Metric `json:"-"`
}

func NewMetric() common.Metric {
	return &Metric{}
}

func CreateMetric(id uint64, name string, enable bool, notify bool, value float64) common.Metric {
	return &Metric{
		ID:     id,
		Name:   name,
		Enable: enable,
		Notify: notify,
		Value:  value}
}

func (metric *Metric) SetID(id uint64) {
	metric.ID = id
}

func (metric *Metric) GetID() uint64 {
	return metric.ID
}

func (metric *Metric) SetDeviceID(id uint64) {
	metric.DeviceID = id
}

func (metric *Metric) GetDeviceID() uint64 {
	return metric.DeviceID
}

func (metric *Metric) SetDataType(datatype int) {
	metric.DataType = datatype
}

func (metric *Metric) GetDataType() int {
	return metric.DataType
}

func (metric *Metric) SetName(name string) {
	metric.Name = name
}

func (metric *Metric) GetName() string {
	return metric.Name
}

func (metric *Metric) SetEnable(enabled bool) {
	metric.Enable = enabled
}

func (metric *Metric) IsEnabled() bool {
	return metric.Enable
}

func (metric *Metric) SetNotify(notify bool) {
	metric.Notify = notify
}

func (metric *Metric) IsNotify() bool {
	return metric.Notify
}

func (metric *Metric) SetKey(key string) {
	metric.Key = key
}

func (metric *Metric) GetKey() string {
	return metric.Key
}

func (metric *Metric) SetUnit(unit string) {
	metric.Unit = unit
}

func (metric *Metric) GetUnit() string {
	return metric.Unit
}

func (metric *Metric) SetAlarmLow(alarm float64) {
	metric.AlarmLow = alarm
}

func (metric *Metric) GetAlarmLow() float64 {
	return metric.AlarmLow
}

func (metric *Metric) SetAlarmHigh(alarm float64) {
	metric.AlarmHigh = alarm
}

func (metric *Metric) GetAlarmHigh() float64 {
	return metric.AlarmHigh
}

func (metric *Metric) SetValue(value float64) {
	metric.Value = value
}

func (metric *Metric) GetValue() float64 {
	return metric.Value
}

func (metric *Metric) SetTimestamp(timestamp *time.Time) {
	metric.Timestamp = timestamp
}

func (metric *Metric) GetTimestamp() *time.Time {
	return metric.Timestamp
}
