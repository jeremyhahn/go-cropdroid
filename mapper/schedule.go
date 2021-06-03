package mapper

type ScheduleMapper interface {
	//MapConfigToModel(schedule config.ScheduleConfig) common.Schedule
	//MapEntityToModel(entity *config.Schedule) config.ScheduleConfig
	//MapModelToEntity(model config.ScheduleConfig) *config.Schedule
}

type DefaultScheduleMapper struct {
}

func NewScheduleMapper() ScheduleMapper {
	return &DefaultScheduleMapper{}
}

/*
func (mapper *DefaultScheduleMapper) MapEntityToModel(scheduleConfig *config.Schedule) config.ScheduleConfig {
	return &config.Schedule{
		ID:        scheduleConfig.GetID(),
		ChannelID: scheduleConfig.GetChannelID(),
		StartDate: scheduleConfig.GetStartDate(),
		EndDate:   scheduleConfig.GetEndDate(),
		Frequency: scheduleConfig.GetFrequency(),
		Interval:  scheduleConfig.GetInterval(),
		Count:     scheduleConfig.GetCount(),
		Days:      scheduleConfig.GetDays()}
}
*/

/*
func (mapper *DefaultScheduleMapper) MapEntityToModel(scheduleConfig *config.Schedule) config.ScheduleConfig {
	return &config.Schedule{
		ID:        scheduleConfig.GetID(),
		ChannelID: scheduleConfig.GetChannelID(),
		StartDate: scheduleConfig.GetStartDate(),
		EndDate:   scheduleConfig.GetEndDate(),
		Frequency: scheduleConfig.GetFrequency(),
		Interval:  scheduleConfig.GetInterval(),
		Count:     scheduleConfig.GetCount(),
		Days:      scheduleConfig.GetDays()}
}

func (mapper *DefaultScheduleMapper) MapModelToEntity(model config.ScheduleConfig) *config.Schedule {
	return &config.Schedule{
		ID:        model.GetID(),
		ChannelID: model.GetChannelID(),
		StartDate: model.GetStartDate(),
		EndDate:   model.GetEndDate(),
		Frequency: model.GetFrequency(),
		Interval:  model.GetInterval(),
		Count:     model.GetCount(),
		Days:      model.GetDays()}
}
*/
