package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrDeviceNotFound = errors.New("Device not found")
)

// FarmStateMap stores real-time metric and channel state for all devices in the farm.
// Note this structure only holds the current state of the farm and its associated
// device states. To access historical data, use the metric datastore.
type FarmStateMap interface {
	SetDevice(deviceType string, deviceState DeviceStateMap)
	GetDevice(deviceType string) (DeviceStateMap, error)
	GetDevices() map[string]DeviceStateMap
	GetMetrics(device string) (map[string]float64, error)
	GetMetricValue(device, key string) (float64, error)
	SetMetricValue(device, key string, value float64) error
	GetChannels(device string) ([]int, error)
	GetChannelValue(device string, channelID int) (int, error)
	SetChannelValue(device string, channelID int, value int) error
	Diff(device string, metrics map[string]float64, channels map[int]int) (DeviceStateDeltaMap, error)
	GetFarmID() uint64
	SetFarmID(uint64)
	GetTimestamp() int64
	//UnmarshalJSON(data []byte) error
	String() string
}

type FarmState struct {
	mutex        *sync.RWMutex             `json:"-"`
	Id           uint64                    `json:"id"`
	Devices      map[string]DeviceStateMap `json:"devices"`
	Timestamp    int64                     `json:"timestamp"`
	FarmStateMap `json:"-"`
}

func NewFarmStateMap(id uint64) FarmStateMap {
	return &FarmState{
		mutex:     &sync.RWMutex{},
		Id:        id,
		Devices:   make(map[string]DeviceStateMap, 0),
		Timestamp: time.Now().Unix()}
}

func CreateFarmState(id uint64, state map[string]DeviceStateMap) FarmStateMap {
	return &FarmState{
		mutex:     &sync.RWMutex{},
		Id:        id,
		Devices:   state,
		Timestamp: time.Now().Unix()}
}

func (farm *FarmState) GetFarmID() uint64 {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Id
}

func (farm *FarmState) SetFarmID(id uint64) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	farm.Id = id
}

func (farm *FarmState) SetDevice(deviceType string, deviceState DeviceStateMap) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	farm.Devices[deviceType] = deviceState
}

func (farm *FarmState) GetDevice(deviceType string) (DeviceStateMap, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if device, ok := farm.Devices[deviceType]; ok {
		return device, nil
	}
	return nil, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d",
		deviceType, farm.Id)
}

func (farm *FarmState) GetDevices() map[string]DeviceStateMap {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Devices
}

func (farm *FarmState) GetMetricValue(device, key string) (float64, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _device, ok := farm.Devices[device]; ok {
		if cachedMetric, ok := _device.GetMetrics()[key]; ok {
			return cachedMetric, nil
		} else {
			return 0.0, fmt.Errorf("Metric not found in farm state: %s.%s", device, key)
		}
	}
	return 0.0, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

func (farm *FarmState) GetMetrics(device string) (map[string]float64, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _device, ok := farm.Devices[device]; ok {
		return _device.GetMetrics(), nil
	}
	return nil, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

func (farm *FarmState) SetMetricValue(device, key string, value float64) error {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	if _device, ok := farm.Devices[device]; ok {
		_device.GetMetrics()[key] = value
		return nil
	}
	return fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

func (farm *FarmState) GetChannelValue(device string, channelID int) (int, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _device, ok := farm.Devices[device]; ok {
		//channels := _device.GetChannels()
		//if channelID > len(channels) || channelID < 0 {
		//	return 0, fmt.Errorf("Invalid channel ID %d (%s device)", channelID, device)
		//}
		return _device.GetChannels()[channelID], nil
	}
	return 0.0, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

func (farm *FarmState) SetChannelValue(device string, channelID int, value int) error {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _device, ok := farm.Devices[device]; ok {
		_device.GetChannels()[channelID] = value
		return nil
	}
	return fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

func (farm *FarmState) GetChannels(device string) ([]int, error) {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	if _device, ok := farm.Devices[device]; ok {
		return _device.GetChannels(), nil
	}
	return nil, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
}

// Diff takes a map of metric and channels that represent a proposed state about to be merged into the state store
// and compares/diffs it against the current stored state. A DeviceStateDeltaMap is returned that contains only the metrics
// and channels that've changed. Metrics take the form map["metric.key"] = float64 and channels map["channel.channelId"] = int.
func (farm *FarmState) Diff(device string, metrics map[string]float64, channels map[int]int) (DeviceStateDeltaMap, error) {
	farm.mutex.Lock()
	defer farm.mutex.Unlock()
	newMetrics := make(map[string]float64, 0)
	newChannels := make(map[int]int, 0)
	if _device, ok := farm.Devices[device]; ok {
		_metrics := _device.GetMetrics()
		_channels := _device.GetChannels()
		for k, v := range metrics {
			if _, ok := _metrics[k]; !ok {
				break
			}
			if metrics[k] == _metrics[k] {
				continue
			}
			_metrics[k] = v
			newMetrics[k] = v
		}
		for k, v := range channels {
			if len(_channels) < k+1 {
				break
			}
			if channels[k] == _channels[k] {
				continue
			}
			_channels[k] = v
			newChannels[k] = v
		}
		if len(newMetrics) == 0 && len(newChannels) == 0 {
			return nil, nil
		}
		return CreateDeviceStateDeltaMap(newMetrics, newChannels), nil
	}
	return nil, fmt.Errorf("Device not found in farm state. device=%s, farmID=%d", device, farm.Id)
	//return CreateDeviceStateDeltaMap(metrics, channels), nil
}

func (farm *FarmState) GetTimestamp() int64 {
	farm.mutex.RLock()
	defer farm.mutex.RUnlock()
	return farm.Timestamp
}

func (farm *FarmState) UnmarshalJSON(data []byte) error {

	if farm.mutex == nil {
		farm.mutex = &sync.RWMutex{}
	}
	farm.mutex.Lock()
	defer farm.mutex.Unlock()

	var message map[string]*json.RawMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return err
	}

	var farmID uint64
	err = json.Unmarshal(*message["id"], &farmID)
	if err != nil {
		return err
	}
	farm.Id = farmID

	var devicesRawMessage map[string]DeviceState
	err = json.Unmarshal(*message["devices"], &devicesRawMessage)
	if err != nil {
		return err
	}

	farm.Devices = make(map[string]DeviceStateMap)
	for k, v := range devicesRawMessage {
		deviceState := v
		farm.Devices[k] = &deviceState
	}

	var timestamp int64
	err = json.Unmarshal(*message["timestamp"], &timestamp)
	if err != nil {
		return err
	}
	farm.Timestamp = timestamp

	return nil
}

func (farm *FarmState) String() string {
	bytes, _ := json.Marshal(farm)
	return string(bytes)
}
