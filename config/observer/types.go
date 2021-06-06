package observer

import (
	"github.com/jeremyhahn/go-cropdroid/config"
)

type FarmConfigObserver interface {
	GetFarmID() int
	//OnFarmChange(farm config.FarmConfig)
	//OnControllerChange(controller config.ControllerConfig)
	OnControllerConfigChange(config config.ControllerConfigConfig)
	OnMetricChange(metric config.MetricConfig)
	OnChannelChange(channel config.ChannelConfig)
	OnConditionChange(condition config.ConditionConfig)
	OnScheduleChange(schedule config.ScheduleConfig)
}
