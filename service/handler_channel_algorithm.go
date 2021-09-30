package service

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/op/go-logging"
)

type ChannelAlgorithmHandler struct {
	logger       *logging.Logger
	service      DeviceService
	device       config.DeviceConfig
	channel      config.ChannelConfig
	metric       config.MetricConfig
	value        float64
	threshold    float64
	backoffTable map[uint64]time.Time
	AlgorithmHandler
}

func NewChannelAlgorithmHandler(logger *logging.Logger, service DeviceService,
	device config.DeviceConfig, channel config.ChannelConfig, metric config.MetricConfig,
	value, threshold float64, backoffTable map[uint64]time.Time) AlgorithmHandler {

	return &ChannelAlgorithmHandler{
		logger:       logger,
		service:      service,
		device:       device,
		channel:      channel,
		metric:       metric,
		value:        value,
		threshold:    threshold,
		backoffTable: backoffTable}
}

func (h *ChannelAlgorithmHandler) Handle() (bool, error) {
	deviceType := h.device.GetType()
	h.logger.Debugf("Processing %s %s algorithm", deviceType, h.channel.GetName())
	if h.channel.GetAlgorithmID() == common.ALGORITHM_PH_ID {
		configs := h.device.GetConfigs()
		gallons := 0
		gallonsConfigKey := fmt.Sprintf("%s.gallons", deviceType)
		for _, config := range configs {
			if config.GetKey() == gallonsConfigKey {
				g, err := strconv.Atoi(config.GetValue())
				if err != nil {
					return false, err
				}
				gallons = g
			}
		}
		if gallons <= 0 {
			return false, fmt.Errorf("%s configuration value must be greater than 0. value: %d", gallonsConfigKey, gallons)
		}
		diff := h.value - h.threshold
		dose := int(math.Round(diff * float64(gallons/2)))
		if dose <= 0 {
			return false, nil
		}
		h.logger.Debugf("Autodosing using pH algorithm: diff=%.2f, dose=%d", diff, dose)
		message := fmt.Sprintf("%s: %.2f, auto-dosing %s for %d seconds", h.metric.GetName(), h.value, h.channel.GetName(), dose)
		_, err := h.service.TimerSwitch(h.channel.GetChannelID(), dose, message)
		if err != nil {
			return false, err
		}
		if h.channel.GetBackoff() > 0 {
			h.backoffTable[h.channel.GetID()] = time.Now()
		}
		return true, nil
	}
	return false, nil
}
