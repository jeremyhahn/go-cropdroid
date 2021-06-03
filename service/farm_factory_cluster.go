// +build cluster

package service

import (
	"fmt"
	"hash/fnv"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/state"

	statemachine "github.com/jeremyhahn/cropdroid/cluster/state"
)

func (fb *FarmFactory) RunClusterProvisionerConsumer() {
	for {
		select {
		case farmConfig := <-fb.farmProvisionerChan:
			fb.app.Logger.Debugf("[FarmFactory.RunClusterProvisionerConsumer] Processing provisioner request...")
			farmConfigChangeChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
			farmStateChangeChan := make(chan state.FarmStateMap, common.BUFFERED_CHANNEL_SIZE)
			farmService, err := fb.BuildClusterService(farmConfig, farmConfigChangeChan, farmStateChangeChan)
			if err != nil {
				fb.app.Logger.Errorf("[FarmFactory.RunClusterProvisionerConsumer] Error: %s", err)
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

	fb.app.Logger.Errorf("[FarmFactory.BuildClusterService] Creating config cluster %d for farm %d",
		farmService.GetConfigClusterID(), farmID)

	if err := fb.createConfigCluster(farmID, farmService.GetConfigClusterID(), farmConfigChangeChan); err != nil {
		fb.app.Logger.Errorf("[FarmFactory.BuildClusterService] Cluster config error: %s", err)
	}

	fb.app.Logger.Errorf("[FarmFactory.BuildClusterService] Creating state cluster: %d", farmID)
	if err := fb.createStateCluster(farmID, farmStateChangeChan); err != nil {
		fb.app.Logger.Errorf("[FarmFactory.BuildClusterService] Cluster state error: %s", err)
	}

	fb.app.RaftCluster.WaitForClusterReady(configClusterID)
	farmService.SetConfig(farmConfig)

	return farmService, nil
}

func (fb *FarmFactory) createConfigCluster(farmID int, clusterID uint64, farmConfigChangeChan chan config.FarmConfig) error {
	if fb.app.RaftCluster != nil {
		params := fb.app.RaftCluster.GetParams()
		params.SetClusterID(clusterID)
		params.SetDataDir(fmt.Sprintf("%s/%d-%d", fb.app.DataDir, farmID, clusterID))
		sm := statemachine.NewFarmConfigMachine(fb.app.Logger, clusterID, farmConfigChangeChan, common.DEFAULT_FARM_CONFIG_HISTORY_LENGTH)
		if err := fb.app.RaftCluster.CreateConfigCluster(clusterID, sm); err != nil {
			return err
		}
	}
	return nil
}

func (fb *FarmFactory) createStateCluster(farmID int, farmStateChangeChan chan state.FarmStateMap) error {
	clusterID := uint64(farmID)
	if fb.app.RaftCluster != nil {
		params := fb.app.RaftCluster.GetParams()
		params.SetClusterID(clusterID)
		params.SetDataDir(fmt.Sprintf("%s/%d", fb.app.DataDir, farmID))
		sm := statemachine.NewFarmStateMachine(fb.app.Logger, clusterID, farmStateChangeChan)
		if err := fb.app.RaftCluster.CreateStateCluster(clusterID, sm); err != nil {
			return err
		}
	}
	return nil
}
