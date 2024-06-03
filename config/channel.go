package config

type CommonChannel interface {
	GetDeviceID() uint64
	SetDeviceID(uint64)
	GetBoardID() int
	SetBoardID(int)
	GetName() string
	SetName(name string)
	IsEnabled() bool
	SetEnable(bool)
	IsNotify() bool
	SetNotify(bool)
	GetDuration() int
	SetDuration(int)
	GetDebounce() int
	SetDebounce(int)
	GetBackoff() int
	SetBackoff(int)
	GetAlgorithmID() uint64
	SetAlgorithmID(uint64)
	KeyValueEntity
}

type Channel interface {
	AddCondition(condition *ConditionStruct)
	GetConditions() []*ConditionStruct
	SetConditions(conditions []*ConditionStruct)
	SetCondition(condition *ConditionStruct)
	GetSchedule() []*ScheduleStruct
	SetSchedule(schedule []*ScheduleStruct)
	SetScheduleItem(schedule *ScheduleStruct)
	CommonChannel
}

type ChannelStruct struct {
	ID          uint64             `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	DeviceID    uint64             `yaml:"device" json:"device_id"`
	BoardID     int                `yaml:"board" json:"board_id"`
	Name        string             `yaml:"name" json:"name"`
	Enable      bool               `yaml:"enable" json:"enable"`
	Notify      bool               `yaml:"notify" json:"notify"`
	Conditions  []*ConditionStruct `gorm:"foreignKey:ChannelID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"conditions" json:"conditions"`
	Schedule    []*ScheduleStruct  `gorm:"foreignKey:ChannelID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" yaml:"schedule" json:"schedule"`
	Duration    int                `yaml:"duration" json:"duration"`
	Debounce    int                `yaml:"debounce" json:"debounce"`
	Backoff     int                `yaml:"backoff" json:"backoff"`
	AlgorithmID uint64             `yaml:"algorithm" json:"algorithm_id"`
	Channel     `sql:"-" gorm:"-" yaml:"-" json:"-"`
}

func NewChannel() *ChannelStruct {
	return &ChannelStruct{
		Conditions: make([]*ConditionStruct, 0),
		Schedule:   make([]*ScheduleStruct, 0)}
}

func (channel *ChannelStruct) TableName() string {
	return "channels"
}

func (channel *ChannelStruct) SetID(id uint64) {
	channel.ID = id
}

func (channel *ChannelStruct) Identifier() uint64 {
	return channel.ID
}

func (channel *ChannelStruct) GetDeviceID() uint64 {
	return channel.DeviceID
}

func (channel *ChannelStruct) SetDeviceID(id uint64) {
	channel.DeviceID = id
}

// Sets the PCB channel ID
func (channel *ChannelStruct) SetBoardID(id int) {
	channel.BoardID = id
}

// Gets the PCB channel ID
func (channel *ChannelStruct) GetBoardID() int {
	return channel.BoardID
}

func (channel *ChannelStruct) SetName(name string) {
	channel.Name = name
}

func (channel *ChannelStruct) GetName() string {
	return channel.Name
}

func (channel *ChannelStruct) SetEnable(enable bool) {
	channel.Enable = enable
}

func (channel *ChannelStruct) IsEnabled() bool {
	return channel.Enable
}

func (channel *ChannelStruct) SetNotify(notify bool) {
	channel.Notify = notify
}

func (channel *ChannelStruct) IsNotify() bool {
	return channel.Notify
}

func (channel *ChannelStruct) SetConditions(conditions []*ConditionStruct) {
	channel.Conditions = conditions
}

func (channel *ChannelStruct) SetCondition(condition *ConditionStruct) {
	for i, c := range channel.Conditions {
		if c.ID == condition.ID {
			channel.Conditions[i] = condition
			return
		}
	}
	channel.Conditions = append(channel.Conditions, condition)
}

func (channel *ChannelStruct) AddCondition(condition *ConditionStruct) {
	channel.Conditions = append(channel.Conditions, condition)
}

func (channel *ChannelStruct) GetConditions() []*ConditionStruct {
	return channel.Conditions
}

func (channel *ChannelStruct) SetSchedule(schedule []*ScheduleStruct) {
	channel.Schedule = schedule
}

func (channel *ChannelStruct) SetScheduleItem(schedule *ScheduleStruct) {
	for i, s := range channel.Schedule {
		if s.ID == schedule.ID {
			channel.Schedule[i] = schedule
			return
		}
	}
	channel.Schedule = append(channel.Schedule, schedule)
}

/*
func (channel *Channel) DeleteSchedule(schedule []Schedule, pos int) []Schedule {
	append(schedule[:pos], schedule[pos+1:]...)
}*/

func (channel *ChannelStruct) GetSchedule() []*ScheduleStruct {
	return channel.Schedule
}

func (channel *ChannelStruct) SetDuration(duration int) {
	channel.Duration = duration
}

func (channel *ChannelStruct) GetDuration() int {
	return channel.Duration
}

func (channel *ChannelStruct) SetDebounce(debounce int) {
	channel.Debounce = debounce
}

func (channel *ChannelStruct) GetDebounce() int {
	return channel.Debounce
}

func (channel *ChannelStruct) SetBackoff(backoff int) {
	channel.Backoff = backoff
}

func (channel *ChannelStruct) GetBackoff() int {
	return channel.Backoff
}

func (channel *ChannelStruct) SetAlgorithmID(id uint64) {
	channel.AlgorithmID = id
}

func (channel *ChannelStruct) GetAlgorithmID() uint64 {
	return channel.AlgorithmID
}
