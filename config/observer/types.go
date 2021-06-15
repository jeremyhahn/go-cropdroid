// +build ignore

package observer

import (
	"github.com/jeremyhahn/go-cropdroid/config"
)

type FarmConfigObserver interface {
	GetFarmID() int
	//OnFarmChange(farm config.FarmConfig)
	//OnDeviceChange(device config.DeviceConfig)
	OnDeviceConfigChange(config config.DeviceConfigConfig)
	OnMetricChange(metric config.MetricConfig)
	OnChannelChange(channel config.ChannelConfig)
	OnConditionChange(condition config.ConditionConfig)
	OnScheduleChange(schedule config.ScheduleConfig)
}
