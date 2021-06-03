package mapper

type MapperRegistry interface {
	GetControllerMapper() ControllerMapper
	GetMetricMapper() MetricMapper
	GetChannelMapper() ChannelMapper
	GetConditionMapper() ConditionMapper
	GetScheduleMapper() ScheduleMapper
	GetUserMapper() UserMapper
}

type MemoryMapperRegistry struct {
	controllerMapper ControllerMapper
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
	controllerMapper := NewControllerMapper(metricMapper, channelMapper)
	return &MemoryMapperRegistry{
		controllerMapper: controllerMapper,
		metricMapper:     metricMapper,
		channelMapper:    channelMapper,
		conditionMapper:  conditionMapper,
		scheduleMapper:   scheduleMapper,
		userMapper:       userMapper}
}

func (registry *MemoryMapperRegistry) GetControllerMapper() ControllerMapper {
	return registry.controllerMapper
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
