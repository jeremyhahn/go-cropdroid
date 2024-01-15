//go:build cluster
// +build cluster

package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jinzhu/gorm"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"

	raftSM "github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
)

type ClusteredFarmFactory struct {
	DefaultFarmFactory
}

type FarmFactoryCluster interface {
	BuildClusterService(farmStateStore state.FarmStorer,
		farmDAO dao.FarmDAO,
		deviceDataStore datastore.DeviceDataStore,
		deviceStateStore state.DeviceStorer,
		farmConfig *config.Farm) (FarmService, error)
	CreateFarmConfigCluster(raftCluster cluster.RaftNode, farmID uint64,
		farmConfigChangeChan chan config.Farm) error
	CreateFarmStateCluster(raftCluster cluster.RaftNode, farmID,
		farmStateID uint64, farmStateChangeChan chan state.FarmStateMap) error
	CreateDeviceStateCluster(
		raftCluster cluster.RaftNode, deviceID uint64, deviceType string,
		deviceStateChangeChan chan common.DeviceStateChange) error
	FarmFactory
}

func NewFarmFactoryCluster(app *app.App, gormDB *gorm.DB,
	datastoreRegistry dao.Registry, serviceRegistry ServiceRegistry,
	deviceMapper mapper.DeviceMapper, changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.Farm, farmTickerProvisionerChan chan uint64,
	farmChannels *FarmChannels) FarmFactoryCluster {

	return &ClusteredFarmFactory{
		DefaultFarmFactory{
			app:              app,
			farmDAO:          datastoreRegistry.GetFarmDAO(),
			deviceDAO:        datastoreRegistry.GetDeviceDAO(),
			deviceSettingDAO: datastoreRegistry.GetDeviceSettingDAO(),
			deviceMapper:     deviceMapper,
			changefeeders:    changefeeders,
			// deviceIndexMap:            make(map[uint64]config.DeviceConfig, 0),
			// channelIndexMap:           make(map[int]config.ChannelConfig, 0),
			// datastoreRegistry:         datastoreRegistry,
			serviceRegistry:           serviceRegistry,
			farmProvisionerChan:       farmProvisionerChan,
			farmTickerProvisionerChan: farmTickerProvisionerChan,
			farmChannels:              farmChannels}}
}

func (cff *ClusteredFarmFactory) BuildClusterService(
	farmStateStore state.FarmStorer,
	farmDAO dao.FarmDAO,
	deviceDataStore datastore.DeviceDataStore,
	deviceStateStore state.DeviceStorer,
	farmConfig *config.Farm) (FarmService, error) {

	clusterServiceRegisty := cff.serviceRegistry.(ClusterServiceRegistry)
	raftCluster := clusterServiceRegisty.GetRaftNode()

	farmID := farmConfig.GetID()

	// TODO: DRY this up with Gossip.Priovision
	farmStateKey := fmt.Sprintf("%s-%d", farmConfig.GetName(), farmID)
	farmStateID := cff.app.IdGenerator.NewID(farmStateKey)

	cff.app.Logger.Debugf("[ClusteredFarmFactory.BuildClusterService] Creating farm, farmID=%d, farmStateID=%d",
		farmID, farmStateID)

	farmConfigChangeChan := cff.farmChannels.FarmConfigChangeChan
	farmStateChangeChan := cff.farmChannels.FarmStateChangeChan
	//deviceStateChangeChan := farmChannels.DeviceStateChangeChan

	// Create config cluster and set initial configuration
	if err := cff.CreateFarmConfigCluster(raftCluster, farmID, farmConfigChangeChan); err != nil {
		cff.app.Logger.Errorf("Cluster config error: %s", err)
	}
	raftCluster.WaitForClusterReady(farmID)

	// Save the Farm config so the device can look it up
	if err := farmDAO.Save(farmConfig); err != nil {
		return nil, err
	}

	// Create device state clusters first so farmService.InitializeState()
	// is able to look up the controllers
	for _, deviceConfig := range farmConfig.GetDevices() {
		deviceID := deviceConfig.GetID()
		deviceType := deviceConfig.GetType()
		if deviceType == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		cff.app.Logger.Errorf("Creating device state cluster for deviceID: %d, farmID: %d",
			deviceConfig.GetID(), farmID)
		if err := cff.CreateDeviceStateCluster(raftCluster, deviceID, deviceType,
			cff.farmChannels.DeviceStateChangeChan); err != nil {
			cff.app.Logger.Errorf("Device state cluster error: %s", err)
		}
		raftCluster.WaitForClusterReady(deviceID)
	}

	// Create state cluster and set initial state
	cff.app.Logger.Errorf("Creating farm state cluster: %d", farmStateID)
	if err := cff.CreateFarmStateCluster(raftCluster, farmID, farmStateID, farmStateChangeChan); err != nil {
		cff.app.Logger.Errorf("Cluster state error: %s", err)
	}
	raftCluster.WaitForClusterReady(farmStateID)

	farmService, err := cff.BuildService(farmStateStore,
		farmDAO, deviceDataStore, deviceStateStore, farmConfig)
	if err != nil {
		return nil, err
	}
	if raftCluster.IsLeader(farmID) {
		cff.app.Logger.Debugf("Setting inital clustered farm config for farm %d", farmID)
		farmService.SetConfig(farmConfig)
	}
	if raftCluster.IsLeader(farmStateID) {
		cff.app.Logger.Debugf("Setting inital clustered farm state for farm %d, farmStateID=%d", farmID, farmStateID)
		farmService.InitializeState(true)
	} else {
		farmService.InitializeState(false)
	}

	return farmService, nil
}

func (cff *ClusteredFarmFactory) CreateFarmConfigCluster(
	raftCluster cluster.RaftNode, farmID uint64,
	farmConfigChangeChan chan config.Farm) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(farmID)
		params.SetDataDir(fmt.Sprintf("%s/%d", cff.app.DataDir, farmID))
		sm := statemachine.NewFarmConfigMachine(cff.app.Logger, cff.app.IdGenerator,
			farmID, cff.app.DataDir, farmConfigChangeChan)
		if err := raftCluster.CreateOnDiskCluster(farmID, params.Join, sm.CreateFarmConfigMachine); err != nil {
			return err
		}
	}
	return nil
}

func (cff *ClusteredFarmFactory) CreateFarmStateCluster(
	raftCluster cluster.RaftNode, farmID, farmStateID uint64,
	farmStateChangeChan chan state.FarmStateMap) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(farmStateID)
		params.SetDataDir(fmt.Sprintf("%s/%d-%d", cff.app.DataDir, farmID, farmStateID))
		sm := statemachine.NewFarmStateMachine(cff.app.Logger, farmStateID, farmStateChangeChan)
		if err := raftCluster.CreateConcurrentCluster(farmStateID, params.Join, sm.CreateFarmStateMachine); err != nil {
			return err
		}
	}
	return nil
}

func (cff *ClusteredFarmFactory) CreateDeviceStateCluster(
	raftCluster cluster.RaftNode, deviceID uint64,
	deviceType string, deviceStateChangeChan chan common.DeviceStateChange) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		params.SetClusterID(deviceID)
		params.SetDataDir(fmt.Sprintf("%s/%d", cff.app.DataDir, deviceID))
		sm := raftSM.NewDeviceStateMachine(cff.app.Logger, deviceID, deviceType, deviceStateChangeChan)
		if err := raftCluster.CreateRegularCluster(deviceID, params.Join, sm.CreateDeviceStateMachine); err != nil {
			return err
		}
	}
	return nil
}
