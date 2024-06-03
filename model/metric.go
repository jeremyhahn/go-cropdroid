package model

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type Metric interface {
	SetValue(value float64)
	GetValue() float64
	config.Metric
}

// The Metric model is a fully populated Metric that contains
// the config, value, and the timestamp the value was last updated.
type MetricStruct struct {
	ID        uint64     `yaml:"id" json:"id"`
	DeviceID  uint64     `yaml:"deviceID" json:"deviceId"`
	DataType  int        `yaml:"datatype" json:"datatype"`
	Name      string     `yaml:"name" json:"name"`
	Key       string     `yaml:"key" json:"key"`
	Enable    bool       `yaml:"enable" json:"enable"`
	Notify    bool       `yaml:"notify" json:"notify"`
	Unit      string     `yaml:"unit" json:"unit"`
	AlarmLow  float64    `yaml:"alarmLow" json:"alarmLow"`
	AlarmHigh float64    `yaml:"alarmHigh" json:"alarmHigh"`
	Value     float64    `yaml:"value" json:"value"`
	Timestamp *time.Time `yaml:"timestamp" json:"timestamp"`
	Metric    `json:"-"`
}

func NewMetric() Metric {
	return &MetricStruct{}
}

func CreateMetric(id uint64, name string, enable bool,
	notify bool, value float64) Metric {

	return &MetricStruct{
		ID:     id,
		Name:   name,
		Enable: enable,
		Notify: notify,
		Value:  value}
}

func (metric *MetricStruct) SetID(id uint64) {
	metric.ID = id
}

func (metric *MetricStruct) Identifier() uint64 {
	return metric.ID
}

func (metric *MetricStruct) SetDeviceID(id uint64) {
	metric.DeviceID = id
}

func (metric *MetricStruct) GetDeviceID() uint64 {
	return metric.DeviceID
}

func (metric *MetricStruct) SetDataType(datatype int) {
	metric.DataType = datatype
}

func (metric *MetricStruct) GetDataType() int {
	return metric.DataType
}

func (metric *MetricStruct) SetName(name string) {
	metric.Name = name
}

func (metric *MetricStruct) GetName() string {
	return metric.Name
}

func (metric *MetricStruct) SetEnable(enabled bool) {
	metric.Enable = enabled
}

func (metric *MetricStruct) IsEnabled() bool {
	return metric.Enable
}

func (metric *MetricStruct) SetNotify(notify bool) {
	metric.Notify = notify
}

func (metric *MetricStruct) IsNotify() bool {
	return metric.Notify
}

func (metric *MetricStruct) SetKey(key string) {
	metric.Key = key
}

func (metric *MetricStruct) GetKey() string {
	return metric.Key
}

func (metric *MetricStruct) SetUnit(unit string) {
	metric.Unit = unit
}

func (metric *MetricStruct) GetUnit() string {
	return metric.Unit
}

func (metric *MetricStruct) SetAlarmLow(alarm float64) {
	metric.AlarmLow = alarm
}

func (metric *MetricStruct) GetAlarmLow() float64 {
	return metric.AlarmLow
}

func (metric *MetricStruct) SetAlarmHigh(alarm float64) {
	metric.AlarmHigh = alarm
}

func (metric *MetricStruct) GetAlarmHigh() float64 {
	return metric.AlarmHigh
}

func (metric *MetricStruct) SetValue(value float64) {
	metric.Value = value
}

func (metric *MetricStruct) GetValue() float64 {
	return metric.Value
}

func (metric *MetricStruct) SetTimestamp(timestamp *time.Time) {
	metric.Timestamp = timestamp
}

func (metric *MetricStruct) GetTimestamp() *time.Time {
	return metric.Timestamp
}
