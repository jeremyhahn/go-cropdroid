// +build cluster

package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/config/store"
	"github.com/jeremyhahn/go-cropdroid/state"
)

func (farm *DefaultFarmService) RunCluster() {

	if farm.running == true {
		farm.app.Logger.Errorf("Farm %d already running!", farm.GetFarmID())
		return
	}
	farm.running = true

	if farm.app.RaftCluster != nil {

		nodeID := farm.app.RaftCluster.GetParams().GetNodeID()
		//clusterID := uint64(farm.GetConfig().GetID())
		farm.app.Logger.Infof("Starting farm. node=%d, farm.id=%d, farm.name=%s", nodeID, farm.GetFarmID())
	}

	go farm.WatchFarmStateChange()
	go farm.WatchFarmConfigChange()

	// Wait for top of the minute
	/*
		func() {
			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ticker.C:
					_, _, secs := time.Now().Clock()
					if secs == 0 {
						ticker.Stop()
						return
					}
					farm.app.Logger.Infof("Waiting for top of the minute... %d sec left", 60-secs)
				}
			}
		}()*/

	farm.pollCluster()
	farm.PollCluster()
}

func (farm *DefaultFarmService) PollCluster() {

	farmConfig, err := farm.configStore.Get(farm.configClusterID)
	if err != nil {
		farm.app.Logger.Errorf("Error looking up farm config %d. Error: ",
			farm.farmID, err.Error())
		return
	}

	if farmConfig == nil {
		farm.app.Logger.Error("farmConfig is null .... Race condition?")
		return
	}

	if farmConfig.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farmConfig.GetInterval()) * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				farm.pollCluster()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}
}

func (farm *DefaultFarmService) pollCluster() {

	farmID := farm.GetFarmID()
	farm.app.Logger.Debugf("Polling farm: %d", farmID)

	if isLeader := farm.app.RaftCluster.WaitForClusterReady(uint64(farmID)); isLeader == false {
		// Only the cluster leader polls the farm
		return
	}

	deviceServices, err := farm.serviceRegistry.GetDeviceServices(farm.GetFarmID())
	if err != nil {
		farm.app.Logger.Errorf("Error: %s", err)
		return
	}
	deltas := make(map[string]state.DeviceStateDeltaMap, len(deviceServices))
	beforeState := farm.GetState()

	for _, device := range deviceServices {

		deviceType := device.GetDeviceType()

		newDeviceState, err := device.Poll()
		if err != nil {
			farm.app.Logger.Errorf("Error polling device state: %s", err)
			return
		}

		farm.app.Logger.Errorf("newDeviceState: %+v", newDeviceState)

		newChannelMap := make(map[int]int, len(newDeviceState.GetChannels()))
		for i, channel := range newDeviceState.GetChannels() {
			newChannelMap[i] = channel
		}

		if beforeState == nil {
			deltas[deviceType] = state.CreateDeviceStateDeltaMap(newDeviceState.GetMetrics(), newChannelMap)
		} else {
			delta, err := farm.state.Diff(deviceType, newDeviceState.GetMetrics(), newChannelMap)
			if err != nil {
				farm.app.Logger.Errorf("Error diffing device state: %s", err)
			}
			if delta != nil {
				deltas[deviceType] = delta
			}
		}

		farm.SetDeviceState(deviceType, newDeviceState)

		if farm.app.MetricDatastore != nil {
			deviceConfig := device.GetDeviceConfig()
			if err := farm.app.MetricDatastore.Save(deviceConfig.GetID(), newDeviceState); err != nil {
				farm.app.Logger.Errorf("Error: %s", err)
				farm.error("Farm.pollCluster", "Farm.pollCluster", err)
				return
			}
		}

		device.Manage()
	}

	farm.app.Logger.Errorf("storing farm state: %s", farm.state)
	if err := farm.stateStore.Put(farmID, farm.state); err != nil {
		farm.app.Logger.Errorf("Error storing farm state: %s", err)
		return
	}

	for deviceType, delta := range deltas {
		farm.PublishDeviceDelta(map[string]state.DeviceStateDeltaMap{deviceType: delta})
	}
}
