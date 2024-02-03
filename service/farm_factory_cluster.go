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

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"

	raftSM "github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
)

type ClusteredFarmFactory struct {
	DefaultFarmFactory
}

type FarmFactoryCluster interface {
	BuildClusterService(farmStateStore state.FarmStorer,
		farmDAO dao.FarmDAO,
		eventLogDAO dao.EventLogDAO,
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
	CreateDeviceDataCluster(raftCluster cluster.RaftNode, deviceDataClusterID uint64) (uint64, error)
	CreateEventLogCluster(raftCluster cluster.RaftNode, eventLogClusterID uint64) (uint64, error)
	FarmFactory
}

func NewFarmFactoryCluster(app *app.App, datastoreRegistry dao.Registry,
	serviceRegistry ServiceRegistry, deviceMapper mapper.DeviceMapper,
	changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.Farm,
	farmTickerProvisionerChan chan uint64,
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
	eventLogDAO dao.EventLogDAO,
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

	// Create EventLog cluster for this farm
	eventLogClusterID, err := cff.CreateEventLogCluster(raftCluster, farmID)
	if err != nil {
		cff.app.Logger.Errorf("Event Log cluster error: %s", err)
	}

	// Create config cluster and set initial configuration
	if err := cff.CreateFarmConfigCluster(raftCluster, farmID, farmConfigChangeChan); err != nil {
		cff.app.Logger.Errorf("Cluster config error: %s", err)
	}

	// Create device state and device data clusters first so farmService.InitializeState()
	// is able to look up the controllers
	devices := farmConfig.GetDevices()
	deviceIds := make([]uint64, 0)
	deviceDataClusterIds := make([]uint64, 0)
	for _, deviceConfig := range devices {
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

		//if farmConfig.GetDataStore() == datastore.RAFT_STORE {
		deviceDataClusterID, err := cff.CreateDeviceDataCluster(raftCluster, deviceID)
		if err != nil {
			cff.app.Logger.Errorf("Device data cluster error: %s", err)
		}

		deviceIds = append(deviceIds, deviceID)
		deviceDataClusterIds = append(deviceDataClusterIds, deviceDataClusterID)
		//}
	}

	// Create farm state cluster and set initial state
	cff.app.Logger.Errorf("Creating farm state cluster: %d", farmStateID)
	if err := cff.CreateFarmStateCluster(raftCluster, farmID, farmStateID, farmStateChangeChan); err != nil {
		cff.app.Logger.Errorf("Cluster state error: %s", err)
	}

	// Wait for all clusters to become ready
	raftCluster.WaitForClusterReady(eventLogClusterID)
	raftCluster.WaitForClusterReady(farmID)
	raftCluster.WaitForClusterReady(farmStateID)
	for i := range deviceIds {
		raftCluster.WaitForClusterReady(deviceIds[i])
		raftCluster.WaitForClusterReady(deviceDataClusterIds[i])
	}

	farmService, err := cff.BuildService(farmStateStore,
		farmDAO, eventLogDAO, deviceDataStore, deviceStateStore, farmConfig)
	if err != nil {
		return nil, err
	}

	// Save the farm config
	if raftCluster.IsLeader(farmID) {
		cff.app.Logger.Debugf("Setting inital clustered farm config for farm %d", farmID)
		farmService.SetConfig(farmConfig) // Farm saved to the database
	}

	// Set the initial farm state
	if raftCluster.IsLeader(farmStateID) {
		cff.app.Logger.Debugf("Setting inital clustered farm state for farm %d, farmStateID=%d", farmID, farmStateID)
		farmService.InitializeState(true) // Farm state saved to the database
	} else {
		farmService.InitializeState(false) // Populate the object but dont save to the database
	}

	// Set the device hardware and firmware versions
	farmService.RefreshHardwareVersions()

	return farmService, nil
}

func (cff *ClusteredFarmFactory) CreateFarmConfigCluster(
	raftCluster cluster.RaftNode, farmID uint64,
	farmConfigChangeChan chan config.Farm) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		sm := statemachine.NewFarmConfigOnDiskStateMachine(cff.app.Logger, cff.app.IdGenerator,
			cff.app.DataDir, farmID, params.NodeID, farmConfigChangeChan)
		if err := raftCluster.CreateOnDiskCluster(farmID, params.Join, sm.CreateFarmConfigOnDiskStateMachine); err != nil {
			return err
		}
	}
	//raftCluster.WaitForClusterReady(farmID)
	return nil
}

func (cff *ClusteredFarmFactory) CreateFarmStateCluster(
	raftCluster cluster.RaftNode, farmID, farmStateID uint64,
	farmStateChangeChan chan state.FarmStateMap) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		sm := statemachine.NewFarmStateConcurrentStateMachine(cff.app.Logger, farmStateID, farmStateChangeChan)
		if err := raftCluster.CreateConcurrentCluster(farmStateID, params.Join, sm.CreateFarmStateConcurrentStateMachine); err != nil {
			return err
		}
	}
	//raftCluster.WaitForClusterReady(farmStateID)
	return nil
}

func (cff *ClusteredFarmFactory) CreateDeviceStateCluster(
	raftCluster cluster.RaftNode, deviceID uint64,
	deviceType string, deviceStateChangeChan chan common.DeviceStateChange) error {

	if raftCluster != nil {
		params := raftCluster.GetParams()
		sm := raftSM.NewDeviceStateConcurrentStateMachine(cff.app.Logger, deviceID, deviceType, deviceStateChangeChan)
		if err := raftCluster.CreateRegularCluster(deviceID, params.Join, sm.CreateDeviceStateConcurrentStateMachine); err != nil {
			return err
		}
	}
	//raftCluster.WaitForClusterReady(deviceID)
	return nil
}

func (cff *ClusteredFarmFactory) CreateDeviceDataCluster(
	raftCluster cluster.RaftNode, deviceID uint64) (uint64, error) {

	deviceDataClusterID := cff.app.IdGenerator.CreateDeviceDataClusterID(deviceID)

	cff.app.Logger.Errorf("Creating device data cluster for deviceDataClusterID: %d, deviceID: %d",
		deviceDataClusterID, deviceID)

	if raftCluster != nil {
		params := raftCluster.GetParams()
		sm := statemachine.NewDeviceDataOnDiskStateMachine(cff.app.Logger, cff.app.IdGenerator,
			cff.app.DataDir, deviceDataClusterID, params.NodeID)
		if err := raftCluster.CreateOnDiskCluster(deviceDataClusterID, params.Join, sm.CreateDeviceDataOnDiskStateMachine); err != nil {
			return 0, err
		}
	}
	//raftCluster.WaitForClusterReady(deviceDataClusterID)
	return deviceDataClusterID, nil
}

func (cff *ClusteredFarmFactory) CreateEventLogCluster(
	raftCluster cluster.RaftNode, farmID uint64) (uint64, error) {

	eventLogClusterID := cff.app.IdGenerator.CreateEventLogClusterID(farmID)

	cff.app.Logger.Errorf("Creating event log cluster: eventLogClusterID: %d, farmID: %d",
		eventLogClusterID, farmID)

	if raftCluster != nil {
		params := raftCluster.GetParams()
		sm := statemachine.NewEventLogOnDiskStateMachine(cff.app.Logger, cff.app.IdGenerator,
			cff.app.DataDir, eventLogClusterID, params.NodeID)
		if err := raftCluster.CreateOnDiskCluster(eventLogClusterID, params.Join, sm.CreateEventLogOnDiskStateMachine); err != nil {
			return 0, err
		}
	}
	//raftCluster.WaitForClusterReady(eventLogClusterID)
	return eventLogClusterID, nil
}
