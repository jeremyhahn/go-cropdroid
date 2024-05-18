//go:build cluster
// +build cluster

package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/state"
)

type ClusteredFarmFactory struct {
	DefaultFarmFactory
}

type FarmFactoryCluster interface {
	BuildClusterService(
		eventLogDAO dao.EventLogDAO,
		farmDAO dao.FarmDAO,
		farmConfig *config.Farm,
		farmStateStore state.FarmStorer,
		deviceStateStore state.DeviceStorer,
		deviceDataStore datastore.DeviceDataStore,
		farmChannels *FarmChannels) (FarmService, error)
	FarmFactory
}

func NewFarmFactoryCluster(app *app.App, datastoreRegistry dao.Registry,
	serviceRegistry ServiceRegistry, deviceMapper mapper.DeviceMapper,
	changefeeders map[string]datastore.Changefeeder,
	farmProvisionerChan chan config.Farm,
	farmTickerProvisionerChan chan uint64) FarmFactoryCluster {

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
			farmTickerProvisionerChan: farmTickerProvisionerChan}}
}

func (cff *ClusteredFarmFactory) BuildClusterService(
	farmEventLogDAO dao.EventLogDAO,
	farmDAO dao.FarmDAO,
	farmConfig *config.Farm,
	farmStateStore state.FarmStorer,
	deviceStateStore state.DeviceStorer,
	deviceDataStore datastore.DeviceDataStore,
	farmChannels *FarmChannels) (FarmService, error) {

	clusterServiceRegisty := cff.serviceRegistry.(ClusterServiceRegistry)
	raftCluster := clusterServiceRegisty.GetRaftNode()

	farmID := farmConfig.ID
	cff.app.Logger.Debugf("BuildClusterService creating clustered farm services for farmID=%d", farmID)

	// Create EventLog cluster for this farm
	if err := farmEventLogDAO.(raft.RaftEventLogDAO).StartClusterNode(false); err != nil {
		cff.app.Logger.Errorf("error starting event log cluster: %s", err)
		return nil, err
	}
	cff.app.Logger.Debugf("Farm Event Log Cluster ID: %s")

	// Create config cluster and set initial configuration
	if err := farmDAO.(raft.RaftFarmConfigDAO).StartClusterNode(farmID, false); err != nil {
		cff.app.Logger.Errorf("error starting farm config cluster: %s", err)
		return nil, err
	}
	cff.app.Logger.Debugf("Farm State Cluster ID: %s")

	// Create device state and data clusters first so farmService.InitializeState() can look them up
	devices := farmConfig.GetDevices()
	deviceIds := make([]uint64, 0)
	deviceDataClusterIds := make([]uint64, 0)
	for _, deviceConfig := range devices {
		deviceID := deviceConfig.ID
		deviceType := deviceConfig.GetType()
		if deviceType == common.CONTROLLER_TYPE_SERVER {
			continue
		}
		if err := deviceStateStore.(raft.RaftDeviceStateStorer).CreateClusterNode(deviceID,
			deviceType, farmChannels.DeviceStateChangeChan); err != nil {
			return nil, err
		}
		deviceDataClusterID, err := deviceDataStore.(raft.RaftDeviceDataDAO).CreateClusterNode(deviceID)
		if err != nil {
			return nil, err
		}
		deviceIds = append(deviceIds, deviceID)
		deviceDataClusterIds = append(deviceDataClusterIds, deviceDataClusterID)

		cff.app.Logger.Debugf("%s device on-disk config cluster ID: %d", deviceType, deviceID)
		cff.app.Logger.Debugf("%s device in-memory state cluster ID: %d", deviceType, deviceID)
		cff.app.Logger.Debugf("%s device timeseries data cluster ID: %d", deviceType, deviceDataClusterID)
	}

	// Create farm state cluster
	farmStateID := farmStateStore.(raft.RaftFarmStateStorer).ClusterID()
	farmStateStore.(raft.RaftFarmStateStorer).StartClusterNode(true)

	// Wait for all clusters to become ready
	raftCluster.WaitForClusterReady(farmEventLogDAO.(raft.RaftEventLogDAO).ClusterID())
	raftCluster.WaitForClusterReady(farmID)
	for i := range deviceIds {
		raftCluster.WaitForClusterReady(deviceIds[i])
		raftCluster.WaitForClusterReady(deviceDataClusterIds[i])
	}

	// Build the FarmService
	farmService, err := cff.BuildService(farmStateStore,
		farmDAO, farmEventLogDAO, deviceDataStore, deviceStateStore, farmConfig, farmChannels)
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

	// Set the device hardware and firmware versions with
	// the current versions published from the device API
	farmService.RefreshHardwareVersions()

	return farmService, nil
}
