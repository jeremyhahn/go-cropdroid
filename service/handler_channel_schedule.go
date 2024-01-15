package service

import (
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/op/go-logging"
)

type ChannelScheduleHandler struct {
	logger          *logging.Logger
	deviceConfig    *config.Device
	channelConfig   *config.Channel
	farmState       state.FarmStateMap
	farmService     FarmService
	deviceService   DeviceService
	scheduleService ScheduleService
	ScheduleHandler
}

func NewChannelScheduleHandler(logger *logging.Logger, deviceConfig *config.Device,
	channelConfig *config.Channel, farmState state.FarmStateMap,
	deviceService DeviceService, scheduleService ScheduleService,
	farmService FarmService) ScheduleHandler {

	return &ChannelScheduleHandler{
		logger:          logger,
		deviceConfig:    deviceConfig,
		channelConfig:   channelConfig,
		farmState:       farmState,
		farmService:     farmService,
		deviceService:   deviceService,
		scheduleService: scheduleService}
}

func (h *ChannelScheduleHandler) Handle() error {

	//eventType := "Scheduled Channel"

	var activeSchedule *config.Schedule

	if !h.channelConfig.IsEnabled() {
		return nil
	}

	deviceState, err := h.farmState.GetDevice(h.deviceConfig.GetType())
	if err != nil {
		return err
	}

	position := deviceState.GetChannels()[h.channelConfig.GetChannelID()]
	h.logger.Debugf("%s switch position: %d", h.channelConfig.GetName(), position)

	for _, schedule := range h.channelConfig.GetSchedule() {

		executionCount := schedule.GetExecutionCount()
		if schedule.GetCount() > 0 && executionCount >= schedule.GetCount() {
			h.logger.Debugf("Reached max execution count: %d", executionCount)
			continue
		}

		if h.scheduleService.IsScheduled(schedule, h.channelConfig.GetDuration()) {
			activeSchedule = schedule
			break
		}
	}

	// if activeSchedule == nil {
	// 	return errors.New("activeSchedule not found")
	// }

	if activeSchedule != nil && activeSchedule.GetID() != 0 {
		h.logger.Debugf("%s scheduled ON condition met. Current position: %d", h.channelConfig.GetName(), position)
		if position == common.SWITCH_OFF {

			message := fmt.Sprintf("Switching ON scheduled %s.", h.channelConfig.GetName())
			_, err := h.deviceService.Switch(h.channelConfig.GetChannelID(), common.SWITCH_ON, message)
			if err != nil {
				return err
			}

			executionCount := activeSchedule.GetExecutionCount() + 1
			activeSchedule.SetLastExecuted(time.Now())
			activeSchedule.SetExecutionCount(executionCount)
			session := CreateSystemSession(h.logger, h.farmService)
			if err := h.scheduleService.Update(session, activeSchedule); err != nil {
				return err
			}
		}

	} else {

		h.logger.Debugf("%s scheduled OFF condition met. Current position: %d", h.channelConfig.GetName(), position)
		if position == common.SWITCH_ON {
			message := fmt.Sprintf("Switching OFF scheduled %s.", h.channelConfig.GetName())
			_, err := h.deviceService.Switch(h.channelConfig.GetChannelID(), common.SWITCH_OFF, message)
			if err != nil {
				return err
			}
		}

	}

	return nil
}
