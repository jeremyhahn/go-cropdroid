package config

type Metric interface {
	GetDeviceID() uint64
	SetDeviceID(uint64)
	GetDataType() int
	SetDataType(int)
	GetKey() string
	SetKey(string)
	GetName() string
	SetName(string)
	IsEnabled() bool
	SetEnable(bool)
	IsNotify() bool
	SetNotify(bool)
	GetUnit() string
	SetUnit(string)
	GetAlarmLow() float64
	SetAlarmLow(float64)
	GetAlarmHigh() float64
	SetAlarmHigh(float64)
	KeyValueEntity
}

type MetricStruct struct {
	ID        uint64  `gorm:"primaryKey" yaml:"id" json:"id"`
	DeviceID  uint64  `yaml:"deviceID" json:"device_id"`
	DataType  int     `gorm:"column:datatype" yaml:"datatype" json:"datatype"`
	Name      string  `yaml:"name" json:"name"`
	Key       string  `yaml:"key" json:"key"`
	Enable    bool    `yaml:"enable" json:"enable"`
	Notify    bool    `yaml:"notify" json:"notify"`
	Unit      string  `yaml:"unit" json:"unit"`
	AlarmLow  float64 `yaml:"alarmLow" json:"alarmLow"`
	AlarmHigh float64 `yaml:"alarmHigh" json:"alarmHigh"`
	Metric    `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewMetric() *MetricStruct {
	return &MetricStruct{}
}

func CreateMetric(id uint64, name string, enable bool,
	notify bool) *MetricStruct {

	return &MetricStruct{
		ID:     id,
		Name:   name,
		Enable: enable,
		Notify: notify}
}

func (metric *MetricStruct) TableName() string {
	return "metrics"
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

func (metric *MetricStruct) GetDataType() int {
	return metric.DataType
}

func (metric *MetricStruct) SetDataType(datatype int) {
	metric.DataType = datatype
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
