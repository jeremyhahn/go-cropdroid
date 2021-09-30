// +build cluster

package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
)

// Watches the DeviceStateChangeChan for real-time incoming device updates.
func (farm *DefaultFarmService) WatchDeviceStateChangeCluster() {

	farm.app.Logger.Debugf("Farm %d watching for incoming device state changes", farm.farmID)

	for {
		select {
		case newDeviceState := <-farm.channels.DeviceStateChangeChan:

			deviceID := newDeviceState.DeviceID
			deviceType := newDeviceState.DeviceType
			stateMap := newDeviceState.StateMap

			_, err := farm.OnDeviceStateChange(deviceType, stateMap)
			if err != nil {
				farm.error("WatchDeviceStateChangeCluster", "WatchDeviceStateChangeCluster", err)
				continue
			}

			if !farm.app.RaftCluster.IsLeader(farm.farmID) {
				continue
			}

			farm.SetDeviceState(deviceType, stateMap)

			if err := farm.deviceDataStore.Save(deviceID, stateMap); err != nil {
				farm.app.Logger.Errorf("Error storing device data: %s", err)
				farm.error("WatchDeviceStateChangeCluster", "WatchDeviceStateChangeCluster", err)
				continue
			}

			deviceService, err := farm.serviceRegistry.GetDeviceService(farm.farmID, deviceType)
			if err != nil {
				farm.app.Logger.Errorf("Error getting device service: %s", err)
				continue
			}

			deviceConfig, err := deviceService.GetConfig()
			if err != nil {
				farm.app.Logger.Errorf("Error getting device config: %s", err)
				continue
			}

			if newDeviceState.IsPollEvent {
				farm.Manage(deviceConfig, farm.GetState())
			}
		}
	}
}

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
	go farm.WatchDeviceStateChangeCluster()

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

	farmConfig, err := farm.configStore.Get(farm.configClusterID, common.CONSISTENCY_CACHED)
	if err != nil || farmConfig == nil {
		farm.app.Logger.Errorf("Farm config not found: configClusterID: %d, farmID: %d",
			farm.configClusterID, farm.farmID)
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

	farm.app.Logger.Debugf("Polling farm, configClusterID=%d, farmID=%d",
		farm.configClusterID, farm.farmID)

	if isLeader := farm.app.RaftCluster.WaitForClusterReady(farm.configClusterID); isLeader == false {
		// Only the cluster leader polls the farm
		return
	}

	farm.poll()
}
