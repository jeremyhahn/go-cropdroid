package common

import "time"

type TimerEvent interface {
	GetChannel() int
	GetDuration() int
	GetTimestamp() time.Time
}

type ChannelTimerEvent struct {
	Channel   int
	Duration  int
	Timestamp time.Time
}

func (event *ChannelTimerEvent) GetChannel() int {
	return event.Channel
}

func (event *ChannelTimerEvent) GetDuration() int {
	return event.Duration
}

func (event *ChannelTimerEvent) GetTimestamp() time.Time {
	return event.Timestamp
}
