package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/op/go-logging"
)

type ChannelConditionHandler struct {
	logger           *logging.Logger
	deviceConfig     config.DeviceConfig
	channelConfig    config.ChannelConfig
	farmState        state.FarmStateMap
	farmService      FarmService
	deviceService    DeviceService
	conditionService ConditionService
	backoffTable     map[uint64]time.Time
	ConditionHandler
}

func NewChannelConditionHandler(logger *logging.Logger, deviceConfig config.DeviceConfig,
	channelConfig config.ChannelConfig, farmState state.FarmStateMap, farmService FarmService,
	deviceService DeviceService, conditionService ConditionService,
	backoffTable map[uint64]time.Time) ConditionHandler {

	return &ChannelConditionHandler{
		logger:           logger,
		deviceConfig:     deviceConfig,
		channelConfig:    channelConfig,
		farmState:        farmState,
		farmService:      farmService,
		deviceService:    deviceService,
		conditionService: conditionService,
		backoffTable:     backoffTable}
}

func (h *ChannelConditionHandler) Handle() (bool, error) {

	var conditionDevice config.DeviceConfig
	var conditionMetric config.MetricConfig
	var condition config.ConditionConfig
	var value float64
	deviceType := h.deviceConfig.GetType()
	result := false
	backoff := h.channelConfig.GetBackoff()
	debounce := h.channelConfig.GetDebounce()

	parse := func(condition config.ConditionConfig) (config.DeviceConfig, config.MetricConfig, error) {
		for _, device := range h.farmService.GetConfig().GetDevices() {
			for _, metric := range device.GetMetrics() {
				if metric.GetID() == condition.GetMetricID() {
					return &device, &metric, nil
				}
			}
		}
		errmsg := fmt.Sprintf("Orphaned condition: %+v", condition)
		h.logger.Error(errmsg)
		return nil, nil, errors.New(errmsg)
	}

	for _, _condition := range h.channelConfig.GetConditions() {

		h.logger.Debugf("Parsing %s condition: %+v", h.channelConfig.GetName(), _condition)

		_conditionDevice, _conditionMetric, err := parse(&_condition)
		if err != nil {
			return false, err
		}
		conditionDevice = _conditionDevice
		conditionMetric = _conditionMetric
		condition = &_condition

		_value, err := h.farmState.GetMetricValue(conditionDevice.GetType(), conditionMetric.GetKey())
		if err != nil {
			return false, err
		}
		value = _value

		_result, err := h.conditionService.IsTrue(&_condition, _value)
		if err != nil {
			return false, err
		}
		result = _result

		if _result {
			break
		}
	}

	position, err := h.farmState.GetChannelValue(deviceType, h.channelConfig.GetChannelID())
	if err != nil {
		return false, err
	}

	if h.channelConfig.GetAlgorithmID() > 0 {
		// Dont continue processing channels managed by algorithms
		handled, err := NewChannelAlgorithmHandler(h.logger, h.deviceService,
			h.deviceConfig, h.channelConfig, conditionMetric, value,
			condition.GetThreshold(), h.backoffTable).Handle()
		if err != nil {
			return false, err
		}
		return handled, nil
	}

	if result {

		h.logger.Debugf("%s conditional evaluated true, current position: %d", h.channelConfig.GetName(), position)

		if position == common.SWITCH_OFF {

			h.logger.Debugf("Switching ON channel: id=%d, name=%s, metric.key=%s, metric.value=%.2f",
				h.channelConfig.GetID(), h.channelConfig.GetName(), conditionMetric.GetKey(), value)

			if h.channelConfig.GetDuration() > 0 {
				message := fmt.Sprintf("Switching ON %s for %d seconds. %s %.2f %s",
					h.channelConfig.GetName(), h.channelConfig.GetDuration(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
				_, err := h.deviceService.TimerSwitch(h.channelConfig.GetChannelID(), h.channelConfig.GetDuration(), message)
				if err != nil {
					return false, err
				}
			} else {
				message := fmt.Sprintf("Switching ON %s. %s %.2f %s", h.channelConfig.GetName(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
				_, err := h.deviceService.Switch(h.channelConfig.GetChannelID(), common.SWITCH_ON, message)
				if err != nil {
					return false, err
				}
			}

			if backoff > 0 {
				h.backoffTable[h.channelConfig.GetID()] = time.Now()
			}

			return true, nil
		}

	} else {

		h.logger.Debugf("%s conditional evaluated false, current position: %d", h.channelConfig.GetName(), position)

		if debounce > 0 {
			h.logger.Debugf("Inspecting debounce window: debounce=%d, threshold=%.2f, value=%.2f",
				debounce, condition.GetThreshold(), value)

			if value >= condition.GetThreshold()-float64(debounce) {
				return false, nil
			}
		}

		if position == common.SWITCH_ON {

			h.logger.Debugf("Switching OFF channel: id=%d, name=%s, metric.key=%s, value=%.2f. debounce=%d",
				h.channelConfig.GetID(), h.channelConfig.GetName(), conditionMetric.GetKey(), value, debounce)

			message := fmt.Sprintf("Switching OFF %s. %s %.2f %s", h.channelConfig.GetName(), conditionMetric.GetName(), value, conditionMetric.GetUnit())
			_, err := h.deviceService.Switch(h.channelConfig.GetChannelID(), common.SWITCH_OFF, message)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil

}
