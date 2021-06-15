// +build cluster

package service

import (
	"fmt"
	"hash/fnv"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"

	statemachine "github.com/jeremyhahn/go-cropdroid/cluster/state"
)

func (fb *FarmFactory) RunClusterProvisionerConsumer() {
	for {
		select {
		case farmConfig := <-fb.farmProvisionerChan:
			fb.app.Logger.Debugf("Processing provisioner request...")
			farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
			farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
			farmService, err := fb.BuildClusterService(farmConfig, farmConfigChangeChan, farmStateChangeChan)
			if err != nil {
				fb.app.Logger.Errorf("Error: %s", err)
				break
			}
			fb.serviceRegistry.AddFarmService(farmService)
			fb.farmTickerProvisionerChan <- farmConfig.GetID()
			go farmService.RunCluster()
		}
	}
}

func (fb *FarmFactory) BuildClusterService(farmConfig config.FarmConfig,
	farmConfigChangeChan chan config.FarmConfig, farmStateChangeChan chan state.FarmStateMap) (FarmService, error) {

	// TODO DRY this up with gossip.Provision
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", farmConfig.GetOrganizationID(), farmConfig.GetID())))
	configClusterID := clusterHash.Sum64()

	farmID := farmConfig.GetID()
	farmService, err := fb.BuildService(farmConfig, farmConfigChangeChan, farmStateChangeChan)
	if err != nil {
		return nil, err
	}

	fb.app.Logger.Debugf("Creating config cluster %d for farm %d",
		farmService.GetConfigClusterID(), farmID)

	if err := fb.createConfigCluster(farmID, farmService.GetConfigClusterID(), farmConfigChangeChan); err != nil {
		fb.app.Logger.Errorf("Cluster config error: %s", err)
	}

	fb.app.Logger.Errorf("Creating state cluster: %d", farmID)
	if err := fb.createStateCluster(farmID, farmStateChangeChan); err != nil {
		fb.app.Logger.Errorf("Cluster state error: %s", err)
	}

	fb.app.RaftCluster.WaitForClusterReady(configClusterID)
	if fb.app.RaftCluster.IsLeader(configClusterID) {
		fb.app.Logger.Debugf("Setting inital clustered farm config for %d from the datastore", configClusterID)
		farmService.SetConfig(farmConfig)
	}

	return farmService, nil
}

func (fb *FarmFactory) createConfigCluster(farmID, configID uint64, farmConfigChangeChan chan config.FarmConfig) error {
	if fb.app.RaftCluster != nil {
		params := fb.app.RaftCluster.GetParams()
		params.SetClusterID(farmID)
		params.SetDataDir(fmt.Sprintf("%s/%d-%d", fb.app.DataDir, farmID, configID))
		sm := statemachine.NewFarmConfigMachine(fb.app.Logger, configID, farmConfigChangeChan, common.DEFAULT_FARM_CONFIG_HISTORY_LENGTH)
		if err := fb.app.RaftCluster.CreateConfigCluster(configID, sm); err != nil {
			return err
		}
	}
	return nil
}

func (fb *FarmFactory) createStateCluster(farmID uint64, farmStateChangeChan chan state.FarmStateMap) error {
	if fb.app.RaftCluster != nil {
		params := fb.app.RaftCluster.GetParams()
		params.SetClusterID(farmID)
		params.SetDataDir(fmt.Sprintf("%s/%d", fb.app.DataDir, farmID))
		sm := statemachine.NewFarmStateMachine(fb.app.Logger, farmID, farmStateChangeChan)
		if err := fb.app.RaftCluster.CreateStateCluster(farmID, sm); err != nil {
			return err
		}
	}
	return nil
}
