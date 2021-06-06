// +build cluster

package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/state"
)

func (farm *DefaultFarmService) RunCluster() {

	if farm.running == true {
		farm.app.Logger.Errorf("[Farm.Run] Farm %d already running!", farm.GetFarmID())
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
	if farm.config.GetInterval() > 0 {
		ticker := time.NewTicker(time.Duration(farm.config.GetInterval()) * time.Second)
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
		// Farm state polling
		return
	}

	controllerServices, err := farm.serviceRegistry.GetControllerServices(farm.GetFarmID())
	if err != nil {
		farm.app.Logger.Errorf("[Farm.pollCluster] Error: %s", err)
		return
	}
	deltas := make(map[string]state.ControllerStateDeltaMap, len(controllerServices))
	beforeState := farm.GetState()

	for _, controller := range controllerServices {

		controllerType := controller.GetControllerType()

		newControllerState, err := controller.Poll()
		if err != nil {
			farm.app.Logger.Errorf("[Farm.pollCluster] Error polling controller state: %s", err)
		}

		farm.app.Logger.Errorf("[Farm.pollCluster] newControllerState: %+v", newControllerState)

		newChannelMap := make(map[int]int, len(newControllerState.GetChannels()))
		for i, channel := range newControllerState.GetChannels() {
			newChannelMap[i] = channel
		}

		if beforeState == nil {
			deltas[controllerType] = state.CreateControllerStateDeltaMap(newControllerState.GetMetrics(), newChannelMap)
		} else {
			delta, err := farm.state.Diff(controllerType, newControllerState.GetMetrics(), newChannelMap)
			if err != nil {
				farm.app.Logger.Errorf("[Farm.pollCluster] Error diffing controller state: %s", err)
			}
			if delta != nil {
				deltas[controllerType] = delta
			}
		}

		farm.SetControllerState(controllerType, newControllerState)

		if farm.app.MetricDatastore != nil {
			if err := farm.app.MetricDatastore.Save(controller.GetControllerConfig().GetID(), newControllerState); err != nil {
				farm.app.Logger.Errorf("[FarmService.pollCluster] Error: %s", err)
				farm.error("Farm.pollCluster", "Farm.pollCluster", err)
				return
			}
		}

		controller.Manage()
	}

	farm.app.Logger.Errorf("[Farm.pollCluster] storing farm state: %s", farm.state)
	if err := farm.stateStore.Put(farmID, farm.state); err != nil {
		farm.app.Logger.Errorf("[Farm.pollCluster] Error storing farm state: %s", err)
		return
	}

	for controllerType, delta := range deltas {
		farm.PublishControllerDelta(map[string]state.ControllerStateDeltaMap{controllerType: delta})
	}
}
