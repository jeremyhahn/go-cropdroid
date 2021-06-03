package config

type Channel struct {
	ID            int         `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	ControllerID  int         `yaml:"controller" json:"controller_id"`
	ChannelID     int         `yaml:"channel" json:"channel_id"`
	Name          string      `yaml:"name" json:"name"`
	Enable        bool        `yaml:"enable" json:"enable"`
	Notify        bool        `yaml:"notify" json:"notify"`
	Conditions    []Condition `yaml:"conditions" json:"conditions"`
	Schedule      []Schedule  `yaml:"schedule" json:"schedule"`
	Duration      int         `yaml:"duration" json:"duration"`
	Debounce      int         `yaml:"debounce" json:"debounce"`
	Backoff       int         `yaml:"backoff" json:"backoff"`
	AlgorithmID   int         `yaml:"algorithm" json:"algorithm_id"`
	ChannelConfig `yaml:"-" json:"-"`
}

func NewChannel() *Channel {
	return &Channel{
		Conditions: make([]Condition, 0),
		Schedule:   make([]Schedule, 0)}
}

func (channel *Channel) SetID(id int) {
	channel.ID = id
}

func (channel *Channel) GetID() int {
	return channel.ID
}

func (channel *Channel) GetControllerID() int {
	return channel.ControllerID
}

func (channel *Channel) SetControllerID(id int) {
	channel.ControllerID = id
}

func (channel *Channel) SetChannelID(id int) {
	channel.ChannelID = id
}

func (channel *Channel) GetChannelID() int {
	return channel.ChannelID
}

func (channel *Channel) SetName(name string) {
	channel.Name = name
}

func (channel *Channel) GetName() string {
	return channel.Name
}

func (channel *Channel) SetEnable(enable bool) {
	channel.Enable = enable
}

func (channel *Channel) IsEnabled() bool {
	return channel.Enable
}

func (channel *Channel) SetNotify(notify bool) {
	channel.Notify = notify
}

func (channel *Channel) IsNotify() bool {
	return channel.Notify
}

func (channel *Channel) SetConditions(conditions []Condition) {
	channel.Conditions = conditions
}

func (channel *Channel) SetCondition(condition ConditionConfig) {
	for i, c := range channel.Conditions {
		if c.GetID() == condition.GetID() {
			channel.Conditions[i] = *condition.(*Condition)
			return
		}
	}
	channel.Conditions = append(channel.Conditions, *condition.(*Condition))
}

func (channel *Channel) AddCondition(condition ConditionConfig) {
	channel.Conditions = append(channel.Conditions, *condition.(*Condition))
}

func (channel *Channel) GetConditions() []Condition {
	return channel.Conditions
}

func (channel *Channel) SetSchedule(schedule []Schedule) {
	channel.Schedule = schedule
}

func (channel *Channel) SetScheduleItem(schedule ScheduleConfig) {
	for i, s := range channel.Schedule {
		if s.GetID() == schedule.GetID() {
			channel.Schedule[i] = *schedule.(*Schedule)
			return
		}
	}
	channel.Schedule = append(channel.Schedule, *schedule.(*Schedule))
}

/*
func (channel *Channel) DeleteSchedule(schedule []Schedule, pos int) []Schedule {
	append(schedule[:pos], schedule[pos+1:]...)
}*/

func (channel *Channel) GetSchedule() []Schedule {
	return channel.Schedule
}

func (channel *Channel) SetDuration(duration int) {
	channel.Duration = duration
}

func (channel *Channel) GetDuration() int {
	return channel.Duration
}

func (channel *Channel) SetDebounce(debounce int) {
	channel.Debounce = debounce
}

func (channel *Channel) GetDebounce() int {
	return channel.Debounce
}

func (channel *Channel) SetBackoff(backoff int) {
	channel.Backoff = backoff
}

func (channel *Channel) GetBackoff() int {
	return channel.Backoff
}

func (channel *Channel) SetAlgorithmID(id int) {
	channel.AlgorithmID = id
}

func (channel *Channel) GetAlgorithmID() int {
	return channel.AlgorithmID
}
