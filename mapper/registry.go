package mapper

type MapperRegistry interface {
	GetDeviceMapper() DeviceMapper
	GetMetricMapper() MetricMapper
	GetChannelMapper() ChannelMapper
	GetConditionMapper() ConditionMapper
	GetScheduleMapper() ScheduleMapper
	GetUserMapper() UserMapper
}

type MemoryMapperRegistry struct {
	deviceMapper DeviceMapper
	metricMapper     MetricMapper
	channelMapper    ChannelMapper
	conditionMapper  ConditionMapper
	scheduleMapper   ScheduleMapper
	userMapper       UserMapper
	MapperRegistry
}

func CreateRegistry() MapperRegistry {
	metricMapper := NewMetricMapper()
	channelMapper := NewChannelMapper()
	conditionMapper := NewConditionMapper()
	scheduleMapper := NewScheduleMapper()
	userMapper := NewUserMapper()
	deviceMapper := NewDeviceMapper(metricMapper, channelMapper)
	return &MemoryMapperRegistry{
		deviceMapper: deviceMapper,
		metricMapper:     metricMapper,
		channelMapper:    channelMapper,
		conditionMapper:  conditionMapper,
		scheduleMapper:   scheduleMapper,
		userMapper:       userMapper}
}

func (registry *MemoryMapperRegistry) GetDeviceMapper() DeviceMapper {
	return registry.deviceMapper
}

func (registry *MemoryMapperRegistry) GetMetricMapper() MetricMapper {
	return registry.metricMapper
}

func (registry *MemoryMapperRegistry) GetChannelMapper() ChannelMapper {
	return registry.channelMapper
}

func (registry *MemoryMapperRegistry) GetConditionMapper() ConditionMapper {
	return registry.conditionMapper
}

func (registry *MemoryMapperRegistry) GetScheduleMapper() ScheduleMapper {
	return registry.scheduleMapper
}

func (registry *MemoryMapperRegistry) GetUserMapper() UserMapper {
	return registry.userMapper
}
