package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormScheduleDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.ScheduleDAO
}

func NewScheduleDAO(logger *logging.Logger, db *gorm.DB) dao.ScheduleDAO {
	return &GormScheduleDAO{logger: logger, db: db}
}

func (dao *GormScheduleDAO) Save(farmID, deviceID uint64, schedule *config.Schedule) error {
	return dao.db.Save(schedule).Error
}

func (dao *GormScheduleDAO) Delete(farmID, deviceID uint64, schedule *config.Schedule) error {
	return dao.db.Delete(schedule).Error
}

func (dao *GormScheduleDAO) GetByChannelID(farmID, deviceID,
	channelID uint64, CONSISTENCY_LEVEL int) ([]*config.Schedule, error) {

	var entities []*config.Schedule
	if err := dao.db.Where("channel_id = ?", channelID).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

/*
func encodeArrayToString(strArray []string) *string {
	days := make([]string, len(strArray))
	for i, day := range strArray {
		days[i] = strings.TrimSpace(day)
	}
	if len(days) == 0 {
		return nil
	}
	str := strings.Join(days, ",")
	return &str
}

func decodeStringToArray(str *string) []string {
	if str == nil {
		return []string{}
	}
	days := strings.Split(*str, ",")
	if len(days) == 0 {
		return []string{}
	}
	return days
}

func (dao *GormScheduleDAO) mapConfigToEntity(schedule config.Schedule) *config.Schedule {
	return &config.Schedule{
		ID:        schedule.GetID(),
		ChannelID: schedule.GetChannelID(),
		StartDate: schedule.GetStartDate(),
		EndDate:   schedule.GetEndDate(),
		Frequency: schedule.GetFrequency(),
		Interval:  schedule.GetInterval(),
		Count:     schedule.GetCount(),
		Days:      schedule.GetDays()}
	//Days:      encodeArrayToString(schedule.GetDays())}
}

func (dao *GormScheduleDAO) mapEntityToConfig(entity *config.Schedule) config.Schedule {
	return &config.Schedule{
		ID: config.GetID(),
		//ChannelID: config.GetChannelID(),
		StartDate: config.GetStartDate(),
		EndDate:   config.GetEndDate(),
		Frequency: config.GetFrequency(),
		Interval:  config.GetInterval(),
		Count:     config.GetCount(),
		Days:      config.GetDays()}
	//Days:      decodeStringToArray(config.GetDays())}
}

/*
func (dao *GormScheduleDAO) Get(id int) (config.Schedule, error) {
	var schedule config.Schedule
	if err := dao.db.First(&schedule, id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (dao *GormScheduleDAO) GetByUserOrgAndDeviceID(orgID, deviceID int) ([]config.Schedule, error) {
	dao.app.Logger.Debugf("Getting schedules for orgID %d and device %d", orgID, deviceID)
	var entities []config.Schedule
	if err := dao.db.Table("schedules").
		Select("schedules.*").
		Joins("JOIN channels on schedules.channel_id = channels.id").
		Joins("JOIN devices on channels.device_id = devices.id AND devices.organization_id = ?", orgID).
		Joins("JOIN permissions on devices.organization_id = permissions.organization_id").
		Where("channels.device_id = ?", deviceID).
		Find(&entities).Error; err != nil {
		return nil, err
	}

	//	schedules := make([]config.Schedule, len(entities))
	//	for i, entity := range entities {
	//		schedules[i] = &entity
	//	}
	//	return schedules, nil

	return entities, nil
}

func (dao *GormScheduleDAO) Update(schedule config.Schedule) error {
	return dao.db.Update(schedule).Error
}

*/
