// +build cluster

package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
)

func (ff *DefaultFarmFactory) BuildClusterService(farmConfig config.FarmConfig) (FarmService, error) {

	//configClusterID := util.ClusterHash(farmConfig.GetOrganizationID(), farmConfig.GetID())

	clusterServiceRegisty := ff.serviceRegistry.(ClusterServiceRegistry)
	raftCluster := clusterServiceRegisty.GetRaftCluster()

	farmID := farmConfig.GetID()
	farmService, err := ff.BuildService(farmConfig)
	if err != nil {
		return nil, err
	}

	configClusterID := farmService.GetConfigClusterID()

	ff.app.Logger.Debugf("Creating farm config cluster %d for farm %d",
		configClusterID, farmID)

	farmChannels := farmService.GetChannels()
	farmConfigChangeChan := farmChannels.FarmConfigChangeChan
	farmStateChangeChan := farmChannels.FarmStateChangeChan
	//deviceStateChangeChan := farmChannels.DeviceStateChangeChan

	// Create config cluster and set initial configuration
	if err := ff.createFarmConfigCluster(raftCluster, farmID, configClusterID, farmConfigChangeChan); err != nil {
		ff.app.Logger.Errorf("Cluster config error: %s", err)
	}
	raftCluster.WaitForClusterReady(configClusterID)
	if raftCluster.IsLeader(configClusterID) {
		ff.app.Logger.Debugf("Setting inital clustered farm config for farm %d, configClusterID=%d", farmID, configClusterID)
		farmService.SetConfig(farmConfig)
	}

	// Create device state clusters first so farmService.InitializeState()
	// is able to look up the controllers
	for _, deviceConfig := range farmConfig.GetDevices() {
		deviceID := deviceConfig.GetID()
		deviceType := deviceConfig.GetType()
		if deviceType == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		ff.app.Logger.Errorf("Creating device state cluster for deviceID: %d, farmID: %d",
			deviceConfig.GetID(), deviceID)
		if err := ff.createDeviceStateCluster(raftCluster, deviceID, deviceType, farmChannels.DeviceStateChangeChan); err != nil {
			ff.app.Logger.Errorf("Device state cluster error: %s", err)
		}
		raftCluster.WaitForClusterReady(deviceID)
	}

	// Create state cluster and set initial state
	ff.app.Logger.Errorf("Creating farm state cluster: %d", farmID)
	if err := ff.createFarmStateCluster(raftCluster, farmID, farmStateChangeChan); err != nil {
		ff.app.Logger.Errorf("Cluster state error: %s", err)
	}
	raftCluster.WaitForClusterReady(farmID)
	if raftCluster.IsLeader(farmID) {
		ff.app.Logger.Debugf("Setting inital farm state for farm %d, configClusterID=%d", farmID, configClusterID)
		farmService.InitializeState(true)
	} else {
		farmService.InitializeState(false)
	}

	return farmService, nil
}

func (ff *DefaultFarmFactory) createFarmConfigCluster(
	raftCluster cluster.RaftCluster, farmID, configID uint64,
	farmConfigChangeChan chan config.FarmConfig) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(configID)
		params.SetDataDir(fmt.Sprintf("%s/%d-%d", ff.app.DataDir, farmID, configID))
		sm := statemachine.NewFarmConfigMachine(ff.app.Logger, configID, farmConfigChangeChan, common.DEFAULT_FARM_CONFIG_HISTORY_LENGTH)
		if err := raftCluster.CreateFarmConfigCluster(configID, sm); err != nil {
			return err
		}
	}
	return nil
}

func (ff *DefaultFarmFactory) createFarmStateCluster(
	raftCluster cluster.RaftCluster, farmID uint64,
	farmStateChangeChan chan state.FarmStateMap) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(farmID)
		params.SetDataDir(fmt.Sprintf("%s/%d", ff.app.DataDir, farmID))
		sm := statemachine.NewFarmStateMachine(ff.app.Logger, farmID, farmStateChangeChan)
		if err := raftCluster.CreateFarmStateCluster(farmID, sm); err != nil {
			return err
		}
	}
	return nil
}

func (ff *DefaultFarmFactory) createDeviceStateCluster(
	raftCluster cluster.RaftCluster, deviceID uint64,
	deviceType string, deviceStateChangeChan chan common.DeviceStateChange) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(deviceID)
		params.SetDataDir(fmt.Sprintf("%s/%d", ff.app.DataDir, deviceID))
		sm := statemachine.NewDeviceStateMachine(ff.app.Logger, deviceID, deviceType, deviceStateChangeChan)
		if err := raftCluster.CreateDeviceStateCluster(deviceID, sm); err != nil {
			return err
		}
	}
	return nil
}
