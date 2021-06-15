// +build cluster

package service

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
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

	farmConfig, err := farm.configStore.Get(farm.farmID, common.CONSISTENCY_CACHED)
	if err != nil {
		farm.app.Logger.Errorf("Error looking up farm config %d. Error: ",
			farm.farmID, err.Error())
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

	if isLeader := farm.app.RaftCluster.WaitForClusterReady(farmID); isLeader == false {
		// Only the cluster leader polls the farm
		return
	}

	farm.poll()
}
