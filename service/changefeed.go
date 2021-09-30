package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type DefaultChangefeedService struct {
	mutex           *sync.Mutex
	app             *app.App
	serviceRegistry ServiceRegistry
	changefeeders   map[string]datastore.Changefeeder
	feeds           map[string]datastore.Changefeeder
	ChangefeedService
}

func NewChangefeedService(app *app.App, serviceRegistry ServiceRegistry,
	changefeeders map[string]datastore.Changefeeder) ChangefeedService {

	service := &DefaultChangefeedService{
		mutex:           &sync.Mutex{},
		app:             app,
		serviceRegistry: serviceRegistry,
		changefeeders:   changefeeders,
		feeds:           changefeeders}

	return service
}

func (service *DefaultChangefeedService) FeedCount() int {
	return len(service.feeds)
}

func (service *DefaultChangefeedService) Subscribe() {
	go service.changefeeders["_device_config_items"].Subscribe(service.OnDeviceConfigConfigChange)
	go service.changefeeders["_channels"].Subscribe(service.OnChannelConfigChange)
	go service.changefeeders["_metrics"].Subscribe(service.OnMetricConfigChange)
	go service.changefeeders["_conditions"].Subscribe(service.OnConditionConfigChange)
	go service.changefeeders["_schedules"].Subscribe(service.OnScheduleConfigChange)
	for k, v := range service.changefeeders {
		if string(k[0]) != "_" {
			if k == "server" {
				continue
			}
			go v.Subscribe(service.OnDeviceStateChange)
		}
	}
}

// OnFarmStateChange handles use cases where the entire farm is updated in one operation, such as in the
// case of a Raft key/value store.
//func (service *DefaultChangefeedService) OnFarmStateChange(farmState state.FarmStateMap) {
//}

// OnDeviceStateChange handles use cases where GORM / MetricDatastore is used to update the farm
// state when a change is made to the underlying database (without FarmService knowing). The database
// must support CDC/changefeed style real-time notifications (cockroachdb / rethinkdb).
func (service *DefaultChangefeedService) OnDeviceStateChange(changefeed datastore.Changefeed) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	record := changefeed.GetRawMessage()

	var deviceID uint64
	err := json.Unmarshal(*record["device_id"], &deviceID)
	if err != nil {
		service.app.Logger.Errorf("Error: %s", err)
		return
	}

	device, ok := service.app.DeviceIndex.Get(deviceID)
	if !ok {
		service.app.Logger.Errorf("Device index missing device: changefeed=%s, device_id=%d",
			changefeed.GetTable(), deviceID)

		service.app.Logger.Warningf("deviceIndex: %+v", service.app.DeviceIndex)
		return
	}

	farmID := device.GetFarmID()
	deviceType := device.GetType()

	farmService := service.serviceRegistry.GetFarmService(farmID)
	if farmService == nil {
		service.app.Logger.Errorf("Service registry missing farm service: farm.id=%d, changefeed=%s, device_id=%f",
			farmID, changefeed.GetTable(), deviceID)
	}

	farmState := farmService.GetState()
	if farmState == nil {
		return
	}

	metricConfigs := device.GetMetrics()
	sort.SliceStable(metricConfigs, func(i, j int) bool {
		return strings.ToLower(metricConfigs[i].GetName()) < strings.ToLower(metricConfigs[j].GetName())
	})
	metrics := make(map[string]float64, 0)
	for _, metric := range metricConfigs {
		var value float64
		jsonKey := strings.ToLower(metric.GetKey())
		err = json.Unmarshal(*record[jsonKey], &value)
		if err != nil {
			service.app.Logger.Errorf("Unmarshal error: %s. changefeed=%s, deviceID=%d",
				err, changefeed.GetTable(), deviceID)
			return
		}
		metricValue, err := farmState.GetMetricValue(deviceType, metric.GetKey())
		if err != nil {
			service.app.Logger.Errorf("Error getting metric value: %s. changefeed=%s, deviceID=%d",
				err, changefeed.GetTable(), deviceID)
			return
		}
		if value == metricValue {
			continue
		}
		metrics[metric.GetKey()] = value
	}

	channelConfigs := device.GetChannels()
	sort.SliceStable(channelConfigs, func(i, j int) bool {
		return strings.ToLower(channelConfigs[i].GetName()) < strings.ToLower(channelConfigs[j].GetName())
	})
	channels := make(map[int]int, 0)
	for i := range channelConfigs {
		var value int
		channel := fmt.Sprintf("c%d", i)
		err = json.Unmarshal(*record[channel], &value)

		channelValue, err := farmState.GetChannelValue(deviceType, i)
		if err != nil {
			service.app.Logger.Errorf("Error retrieving metric from farm state: changefeed=%s, device_id=%f, error=%s",
				changefeed.GetTable(), deviceID, err)
			return
		}
		if value == channelValue {
			continue
		}

		channels[channelConfigs[i].GetChannelID()] = value
	}

	deviceStateDelta, err := farmState.Diff(deviceType, metrics, channels)
	if err != nil {
		service.app.Logger.Errorf("Error: %s. changefeed=%s, deviceID=%d",
			err, changefeed.GetTable(), deviceID)
		return
	}

	if deviceStateDelta == nil {
		service.app.Logger.Debugf("No state difference, aborting. changefeed=%s, device_id=%d",
			changefeed.GetTable(), deviceID)
		return
	}

	param := make(map[string]state.DeviceStateDeltaMap, 1)
	param[deviceType] = deviceStateDelta

	farmService.PublishDeviceDelta(param)
}

func (service *DefaultChangefeedService) OnDeviceConfigConfigChange(changefeed datastore.Changefeed) {

	service.mutex.Lock()
	defer service.mutex.Unlock()

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.id: %s", err)
		return
	}

	var deviceID uint64
	err = json.Unmarshal(*record["device_id"], &deviceID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.deviceID: %s", err)
		return
	}

	var userID uint64
	err = json.Unmarshal(*record["user_id"], &userID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.userID: %s", err)
		return
	}

	var deviceConfigItem config.DeviceConfigItem
	err = json.Unmarshal(changefeed.GetBytes(), &deviceConfigItem)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling changfeed: %s", err)
	}
	deviceConfigItem.SetID(id)
	deviceConfigItem.SetUserID(userID)
	deviceConfigItem.SetDeviceID(deviceID)

	service.app.Logger.Errorf("OnDeviceConfigConfigChange fired! Received device config: %+v", &deviceConfigItem)

	device, ok := service.app.DeviceIndex.Get(deviceID)
	if !ok {
		service.app.Logger.Errorf("Device index missing device: changefeed=%s, device_id=%d, index=%+v",
			changefeed.GetTable(), deviceID, service.app.DeviceIndex.GetAll())
		return
	}
	device.SetConfig(&deviceConfigItem)

	if farmService := service.serviceRegistry.GetFarmService(device.GetFarmID()); farmService != nil {
		farmConfig := farmService.GetConfig()
		farmConfig.ParseConfigs()
		farmService.PublishConfig(farmConfig)
	}
}

func (service *DefaultChangefeedService) OnMetricConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id int
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.id: %s", err)
		return
	}

	var deviceID uint64
	err = json.Unmarshal(*record["device_id"], &deviceID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.deviceID: %s", err)
		return
	}

	var metric config.Metric
	err = json.Unmarshal(changefeed.GetBytes(), &metric)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling changfeed: %s", err)
	}
	metric.SetID(id)
	metric.SetDeviceID(deviceID)

	service.app.Logger.Errorf("OnMetricChange fired! Received metric config: %+v", metric)

	device, ok := service.app.DeviceIndex.Get(metric.GetDeviceID())
	if !ok {
		return
	}
	device.SetMetric(&metric)

	if farmService := service.serviceRegistry.GetFarmService(device.GetFarmID()); farmService != nil {
		farmConfig := farmService.GetConfig()
		farmConfig.ParseConfigs()
		farmService.PublishConfig(farmConfig)
	}
}

func (service *DefaultChangefeedService) OnChannelConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.id: %s", err)
		return
	}

	var deviceID uint64
	err = json.Unmarshal(*record["device_id"], &deviceID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.deviceID: %s", err)
		return
	}

	var algorithmID int
	err = json.Unmarshal(*record["algorithm_id"], &deviceID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.algorithmID: %s", err)
		return
	}

	var channel config.Channel
	err = json.Unmarshal(changefeed.GetBytes(), &channel)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling changfeed: %s", err)
	}
	channel.SetID(id)
	channel.SetDeviceID(deviceID)
	channel.SetAlgorithmID(algorithmID)

	service.app.Logger.Errorf("OnChannelConfigChange fired! Received channel config: %+v", channel)

	device, ok := service.app.DeviceIndex.Get(deviceID)
	if !ok {
		service.app.Logger.Errorf("OnChannelConfigChange: Failed to retrieve device from index: %+v", service.app.DeviceIndex)
		return
	}
	device.SetChannel(&channel)

	if farmService := service.serviceRegistry.GetFarmService(device.GetFarmID()); farmService != nil {
		farmConfig := farmService.GetConfig()
		farmConfig.ParseConfigs()
		farmService.PublishConfig(farmConfig)
	}
}

func (service *DefaultChangefeedService) OnConditionConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.id: %s", err)
		return
	}

	var channelID uint64
	err = json.Unmarshal(*record["channel_id"], &channelID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.channelID: %s", err)
		return
	}

	var metricID int
	err = json.Unmarshal(*record["metric_id"], &metricID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.metricID: %s", err)
		return
	}

	var condition config.Condition
	err = json.Unmarshal(changefeed.GetBytes(), &condition)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling changfeed: %s", err)
	}
	condition.SetID(id)
	condition.SetChannelID(channelID)
	condition.SetMetricID(metricID)

	service.app.Logger.Errorf("OnConditionConfigChange fired! Received condition config: %+v", condition)

	channel, ok := service.app.ChannelIndex.Get(channelID)
	if !ok {
		return
	}

	channel.SetCondition(&condition)

	device, ok := service.app.DeviceIndex.Get(channel.GetDeviceID())
	if !ok {
		return
	}

	if farmService := service.serviceRegistry.GetFarmService(device.GetFarmID()); farmService != nil {
		farmConfig := farmService.GetConfig()
		farmConfig.ParseConfigs()
		farmService.PublishConfig(farmConfig)
	}
}

func (service *DefaultChangefeedService) OnScheduleConfigChange(changefeed datastore.Changefeed) {

	record := changefeed.GetRawMessage()

	var id uint64
	err := json.Unmarshal(*record["id"], &id)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.id: %s", err)
		return
	}

	var channelID uint64
	err = json.Unmarshal(*record["channel_id"], &channelID)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling channel.channelID: %s", err)
		return
	}

	var schedule config.Schedule
	err = json.Unmarshal(changefeed.GetBytes(), &schedule)
	if err != nil {
		service.app.Logger.Errorf("Error unmarshaling changfeed: %s", err)
	}
	schedule.SetID(id)
	schedule.SetChannelID(channelID)

	service.app.Logger.Errorf("OnScheduleConfigChange fired! Received condition config: %+v", schedule)

	channel, ok := service.app.ChannelIndex.Get(channelID)
	if !ok {
		return
	}

	channel.SetScheduleItem(&schedule)

	device, ok := service.app.DeviceIndex.Get(channel.GetDeviceID())
	if !ok {
		return
	}

	if farmService := service.serviceRegistry.GetFarmService(device.GetFarmID()); farmService != nil {
		farmConfig := farmService.GetConfig()
		farmConfig.ParseConfigs()
		farmService.PublishConfig(farmConfig)
	}
}
