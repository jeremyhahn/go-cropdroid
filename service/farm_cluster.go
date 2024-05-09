//go:build cluster
// +build cluster

package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
)

// Watches the DeviceStateChangeChan for real-time incoming device updates.
func (farm *DefaultFarmService) WatchDeviceStateChangeCluster() {

	farm.app.Logger.Debugf("Farm %d watching for incoming device state changes", farm.farmID)

	clusterServiceRegisty := farm.serviceRegistry.(ClusterServiceRegistry)
	raftCluster := clusterServiceRegisty.GetRaftNode()

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

			if !raftCluster.IsLeader(farm.farmID) {
				farm.app.Logger.Debugf("Not the raft leader for farmID=%d, ignoring DeviceStateChange event.", farm.farmID)
				continue
			}

			farm.SetDeviceState(deviceType, stateMap)

			if err := farm.deviceDataStore.Save(deviceID, stateMap); err != nil {
				farm.app.Logger.Errorf("Error storing device data: %s", err)
				farm.error("WatchDeviceStateChangeCluster", "WatchDeviceStateChangeCluster", err)
				continue
			}

			if newDeviceState.IsPollEvent {
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
				farm.Manage(deviceConfig, farm.GetState())
			}
		case <-farm.deviceStateQuitChan:
			farm.app.Logger.Debugf("Closing device state channel. farmID=%d", farm.farmID)
			return
		}
	}
}

func (farm *DefaultFarmService) RunCluster() {

	if farm.running == true {
		farm.app.Logger.Errorf("Farm %d already running!", farm.GetFarmID())
		return
	}
	farm.running = true

	clusterServiceRegisty := farm.serviceRegistry.(ClusterServiceRegistry)
	raftCluster := clusterServiceRegisty.GetRaftNode()

	if raftCluster != nil {

		nodeID := raftCluster.GetParams().GetNodeID()
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

	farm.pollCluster(raftCluster)
	farm.PollCluster(raftCluster)
}

func (farm *DefaultFarmService) PollCluster(raftCluster cluster.RaftNode) {

	farmConfig, err := farm.farmDAO.Get(farm.farmID, common.CONSISTENCY_LOCAL)
	if err != nil || farmConfig.ID == 0 {
		farm.app.Logger.Errorf("Farm config not found: farmID: %d", farm.farmID)
		return
	}

	if farmConfig.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farmConfig.GetInterval()) * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				farm.pollCluster(raftCluster)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}
}

func (farm *DefaultFarmService) pollCluster(raftCluster cluster.RaftNode) {
	farm.app.Logger.Debugf("Polling clustered farm %d", farm.farmID)
	if isLeader := raftCluster.WaitForClusterReady(farm.farmID); isLeader == false {
		farm.app.Logger.Warningf("Aborting polling, not the cluster leader for farm: %d", farm.farmID)
		// Only the cluster leader polls the farm
		return
	}
	farm.poll()
}
