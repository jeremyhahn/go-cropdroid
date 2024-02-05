//go:build !cluster
// +build !cluster

package service

import (
	"strings"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

const (
	DOES_NOT_REPEAT = "DOES NOT REPEAT"
	DAILY           = "DAILY"
	WEEKLY          = "WEEKLY"
	MONTHLY         = "MONTHLY"
	YEARLY          = "YEARLY"
	CUSTOM          = "CUSTOM..."
)

type ScheduleService interface {
	GetNow() *time.Time
	GetSchedule(session Session, channelID uint64) ([]*config.Schedule, error)
	//GetSchedules(user common.UserAccount, deviceID int) ([]config.Schedule, error)
	Create(session Session, schedule *config.Schedule) (*config.Schedule, error)
	Update(session Session, schedule *config.Schedule) error
	Delete(session Session, schedule *config.Schedule) error
	IsScheduled(schedule *config.Schedule, duration int) bool
}

type DefaultScheduleService struct {
	app           *app.App
	dao           dao.ScheduleDAO
	mapper        mapper.ScheduleMapper
	now           *time.Time
	configService ConfigService
	ScheduleService
}

// NewScheduleService creates a new default ScheduleService instance using the current time as "now"
func NewScheduleService(app *app.App, scheduleDAO dao.ScheduleDAO,
	scheduleMapper mapper.ScheduleMapper,
	configService ConfigService) ScheduleService {

	return &DefaultScheduleService{
		app:           app,
		dao:           scheduleDAO,
		mapper:        scheduleMapper,
		now:           nil,
		configService: configService}
}

// CreateScheduleService creates a new ScheduleService instance using the specified time as "now"
func CreateScheduleService(app *app.App, scheduleDAO dao.ScheduleDAO,
	scheduleMapper mapper.ScheduleMapper, now time.Time,
	configService ConfigService) (ScheduleService, error) {

	app.Logger.Debugf("Setting current time to %s", now)
	return &DefaultScheduleService{
		app:           app,
		dao:           scheduleDAO,
		mapper:        scheduleMapper,
		now:           &now,
		configService: configService}, nil
}

// GetNow returns the current time used for calculations within the ScheduleService
func (service *DefaultScheduleService) GetNow() *time.Time {
	if service.now != nil {
		return service.now
	}
	now := time.Now().In(service.app.Location)
	nowHr, nowMin, nowSec := now.Clock()
	roundedToSecond := time.Date(now.Year(), now.Month(),
		now.Day(), nowHr, nowMin, nowSec, 0, service.app.Location)
	return &roundedToSecond
}

// GetSchedule retrieves a specific schedule entry from the database
func (service *DefaultScheduleService) GetSchedule(session Session,
	channelID uint64) ([]*config.Schedule, error) {
	farmID := session.GetRequestedFarmID()
	deviceID := uint64(0)
	consistencyLeveL := session.GetFarmService().GetConsistencyLevel()
	entities, err := service.dao.GetByChannelID(farmID, deviceID,
		channelID, consistencyLeveL)
	if err != nil {
		return nil, err
	}
	schedules := make([]*config.Schedule, len(entities))
	for i, entity := range entities {
		//schedules[i] = service.mapper.MapEntityToModel(&entity)
		schedules[i] = entity
	}
	return schedules, nil
	//return entities, nil
}

/*
// GetSchedules retrieves a list of schedule entries from the database
func (service *DefaultScheduleService) GetSchedules(user common.UserAccount, deviceID int) ([]config.Schedule, error) {
	entities, err := service.dao.GetByUserOrgAndDeviceID(user.GetOrganizationID(), deviceID)
	if err != nil {
		return nil, err
	}
	schedules := make([]config.Schedule, len(entities))
	for i, entity := range entities {
		schedules[i] = service.mapper.MapEntityToModel(&entity)
	}
	return schedules, nil
}
*/

// Create a new schedule entry
func (service *DefaultScheduleService) Create(session Session, schedule *config.Schedule) (*config.Schedule, error) {
	// v0.0.3a: farmService.SetDeviceConfig updates the entity now
	//entity := service.mapper.MapModelToEntity(schedule)
	// if err := service.dao.Create(schedule); err != nil {
	// 	return nil, err
	// }
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			if channel.GetID() == schedule.GetChannelID() {
				channel.SetScheduleItem(schedule)
				device.SetChannel(channel)
				return schedule, farmService.SetDeviceConfig(device)
			}
		}
	}
	return nil, ErrScheduleNotFound
}

// Update an existing schedule entry in the database
func (service *DefaultScheduleService) Update(session Session, schedule *config.Schedule) error {
	// v0.0.3a: farmService.SetDeviceConfig updates the entity now
	// if err := service.dao.Save(schedule); err != nil {
	// 	return err
	// }
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			for _, schedule := range channel.GetSchedule() {
				if channel.GetID() == schedule.GetChannelID() {
					channel.SetScheduleItem(schedule)
					device.SetChannel(channel)
					return farmService.SetDeviceConfig(device)
				}
			}
		}
	}
	return ErrScheduleNotFound
}

// Delete a schedule entry from the database
func (service *DefaultScheduleService) Delete(session Session, schedule *config.Schedule) error {
	// v0.0.3a: farmService.SetDeviceConfig doesnt delete the schedule :(
	farmID := session.GetRequestedFarmID()
	deviceID := uint64(0)
	if err := service.dao.Delete(farmID, deviceID, schedule); err != nil {
		return err
	}
	farmService := session.GetFarmService()
	farmConfig := farmService.GetConfig()
	for _, device := range farmConfig.GetDevices() {
		for _, channel := range device.GetChannels() {
			// Android client only sends the schedule id on delete
			//if channel.GetChannelID() == schedule.GetChannelID() {
			for i, _schedule := range channel.GetSchedule() {
				if _schedule.GetID() == schedule.GetID() {
					channel.Schedule = append(channel.Schedule[:i], channel.Schedule[i+1:]...)
					device.SetChannel(channel)
					return farmService.SetDeviceConfig(device)
				}
			}
			//}
		}
	}
	return ErrScheduleNotFound
}

// IsScheduled takes a Schedule and number of seconds that specify the duration of time the switch should be on
func (service *DefaultScheduleService) IsScheduled(schedule *config.Schedule, duration int) bool {
	service.app.Logger.Debugf("schedule=%+v", schedule)
	startDate := schedule.GetStartDate()
	if schedule.GetFrequency() > 0 {
		now := time.Now().In(service.app.Location)

		if schedule.GetDays() != nil {
			days := strings.Split(*schedule.GetDays(), ",")
			//if len(schedule.GetDays()) > 0 {
			if len(days) > 0 {
				pieces := strings.Split(schedule.GetStartDate().Format(common.TIME_RFC1123_FORMAT), ",")
				today := strings.ToUpper(pieces[0][:2])
				var isToday = false
				for _, day := range days {
					if day == today {
						isToday = true
						break
					}
				}
				if !isToday {
					return false
				}
			}
		}

		switch schedule.GetFrequency() {

		case common.SCHEDULE_FREQUENCY_DAILY:
			if schedule.GetInterval() > 0 {
				if now.Day()-startDate.Day() < schedule.GetInterval() {
					return false
				}
			}

		case common.SCHEDULE_FREQUENCY_WEEKLY:
			if schedule.GetInterval() > 0 {
				if now.Day()-startDate.Day() < 7*schedule.GetInterval() {
					return false
				}
			}
			if now.Day()-startDate.Day() != 7 {
				return false
			}

		case common.SCHEDULE_FREQUENCY_MONTHLY:
			if schedule.GetInterval() > 0 {
				if int(now.Month())-int(startDate.Month()) < schedule.GetInterval() {
					return false
				}
			}
			if now.Day() != startDate.Day() {
				return false
			}

		case common.SCHEDULE_FREQUENCY_YEARLY:
			if schedule.GetInterval() > 0 {
				if int(now.Year())-int(startDate.Year()) < schedule.GetInterval() {
					return false
				}
			}
			if now.Month() != startDate.Month() || now.Day() != startDate.Day() {
				return false
			}
		}
	}
	if duration > 0 {
		timerExpiration := service.timeWithDuration(startDate, duration)
		return service.isTimeBetween(startDate, timerExpiration)
	}
	endDate := schedule.GetEndDate()
	if endDate == nil {
		return true // No timer duration or end date - run forever
	}
	return service.isDateTimeBetween(startDate, *endDate)
}

// Compares a start and end time in HH:MM 24 hr format against the current time and returns true if
// its between the specified start and end time.
func (service *DefaultScheduleService) isTimeBetween(startDate, endDate time.Time) bool {

	now := service.GetNow()
	nowHr, nowMin, _ := now.Clock()
	startHr, startMin, _ := startDate.Clock()
	endHr, endMin, _ := endDate.Clock()

	service.app.Logger.Debugf("[ScheduleService.isTimeBetween] startHr=%d,startMin=%d,endHr=%d,endMin=%d,nowHour=%d,nowMinute=%d",
		startHr, startMin, endHr, endMin, nowHr, nowMin)

	if startHr == endHr {
		if nowHr != startHr {
			return false
		}
		return service.isScheduledMinute(startMin, endMin, nowMin)
	}

	if nowHr >= startHr || nowHr <= endHr {
		if nowHr != startHr && nowHr != endHr && startHr != endHr {
			return true
		}
		if service.isScheduledMinute(startMin, endMin, nowMin) || (endMin == 0 && endHr != nowHr) {
			return true
		}
	}
	return false
}

// isBetween returns true if "now" date and time is between the specified start and end date and time
func (service *DefaultScheduleService) isDateTimeBetween(startDate, endDate time.Time) bool {
	now := service.GetNow()
	startDate = startDate.In(service.app.Location)
	endDate = endDate.In(service.app.Location)
	service.app.Logger.Debugf("[ScheduleService.isNowBetween] Comparing schedule: now=%s, start=%s, stop=%s", now, startDate, endDate)
	return now.After(startDate) && now.Before(endDate) || now.Equal(startDate) // || now.Equal(endDate)
}

// timeWithDuration returns the specified time with the configured timer duration added
func (service *DefaultScheduleService) timeWithDuration(t time.Time, duration int) time.Time {
	timeWithDuration := t.Add(time.Duration(duration) * time.Second)
	service.app.Logger.Debugf("[ScheduleService.timeWithDuration] time=%s, timeWithDuration=%s, duration=%d", t, timeWithDuration, duration)
	return timeWithDuration
}

// IsScheduledMinute returns true if scheduled at the current minute
func (service *DefaultScheduleService) isScheduledMinute(start, end, nowMin int) bool {
	service.app.Logger.Debugf("[isScheduledMinute] start=%d,end=%d,nowMin=%d", start, end, nowMin)
	//return nowMin >= start && nowMin < end
	if start == end {
		return nowMin >= start
	}
	return nowMin >= start && nowMin < end
}
