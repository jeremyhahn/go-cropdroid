//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/app"
	clusterutil "github.com/jeremyhahn/go-cropdroid/cluster/util"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"

	"github.com/jeremyhahn/go-cropdroid/cluster/statemachine"
)

const (
	TestSuiteName         = "datastore_raft"
	TestDataDir           = "test-data"
	ClusterID             = uint64(420) // system / gossip raft
	OrganizationClusterID = uint64(100)
	FarmConfigClusterID   = uint64(101)
	FarmStateClusterID    = uint64(102)
	DeviceConfigClusterID = uint64(103)
	UserClusterID         = uint64(104)
	RoleClusterID         = uint64(105)
	AlgorithmClusterID    = uint64(106)
	RegistrationClusterID = uint64(107)
	DeviceStateClusterID  = uint64(108)
	DeviceDataClusterID   = uint64(109)
	CustomerClusterID     = uint64(110)
	NodeCount             = 3
	RaftLeaderID          = 3
	NodeID                = 1
)

var (
	ConcurrentTestCounter              = 0
	Cluster               *TestCluster = nil
)

type TestCluster struct {
	app         *app.App
	nodeCount   int
	gossipMutex sync.Mutex
	gossipNodes []GossipNode
	raftMutex   sync.Mutex
	raftNodes   []RaftNode
	clusterID   uint64
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	Cluster.Cleanup()

	Cluster := NewClusterIntegrationTest()
	Cluster.StartCluster()
}

func teardown() {
	// if CurrentTest != nil {
	// 	CurrentTest.Cleanup()
	// }
}

func NewClusterIntegrationTest() *TestCluster {

	if NodeCount < 3 {
		panic("NodeCount must be greater than 3")
	}

	if NodeCount > 9 {
		panic("NodeCount must be less than 9")
	}

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	app := &app.App{
		Logger:      logger,
		Location:    location,
		NodeID:      NodeID,
		ClusterID:   ClusterID,
		DataDir:     fmt.Sprintf("./%s", TestDataDir),
		IdGenerator: idGenerator,
		IdSetter:    util.NewIdSetter(idGenerator)}

	Cluster = &TestCluster{
		app:       app,
		nodeCount: NodeCount,
		clusterID: ClusterID}

	return Cluster
}

func (dt *TestCluster) StartCluster() {

	if ConcurrentTestCounter > 0 {
		ConcurrentTestCounter++
		return
	}

	clusterIaasProvider := ""
	clusterRegion := ""
	clusterZone := ""
	localIP := util.ParseLocalIP()
	gossipStartPort := 60010
	raftStartPort := 60020

	gossipPeers := dt.createPeers(localIP, gossipStartPort)
	raftPeers := dt.createPeers(localIP, raftStartPort)

	localAddress := ""
	clusterJoin := false
	clusterVirtualNodes := 1
	clusterMaxNodes := 7
	clusterBootstrap := 3
	clusterInit := false

	farmProvisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmDeprovisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
	farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)

	var wg sync.WaitGroup

	for i := 0; i < dt.nodeCount; i++ {

		nodeID := uint64(i)
		clusterGossipPort := gossipStartPort + i
		clusterRaftPort := raftStartPort + i

		raftOptions := clusterutil.RaftOptions{
			Port:              clusterRaftPort,
			RequestedLeaderID: RaftLeaderID,
			SystemClusterID:   ClusterID,
			//PermissionClusterID: PermissionClusterID,
			UserClusterID: UserClusterID,
			RoleClusterID: RoleClusterID}

		params := clusterutil.NewClusterParams(dt.app.Logger, raftOptions, nodeID,
			clusterIaasProvider, clusterRegion, clusterZone, dt.app.DataDir, localAddress,
			localIP, gossipPeers, raftPeers, clusterJoin, clusterGossipPort,
			clusterRaftPort, RaftLeaderID, clusterVirtualNodes, clusterMaxNodes,
			clusterBootstrap, clusterInit, dt.app.IdGenerator, dt.app.IdSetter,
			farmProvisionerChan, farmDeprovisionerChan, farmTickerProvisionerChan)

		wg.Add(1)
		go func(params *clusterutil.ClusterParams) {

			gossipNode := NewGossipNode(params, clusterutil.NewHashring(clusterVirtualNodes))
			gossipNode.Join()
			go gossipNode.Run()

			// Need to pass in a new cluster ID
			// raftNode := cluster.NewRaftNode(params, util.NewHashring(1))
			// for raftNode == nil {
			// 	dt.app.Logger.Info("Waiting for enough nodes to build the Raft quorum...")
			// 	time.Sleep(1 * time.Second)
			// 	raftNode = gossipNode.GetSystemRaft()
			// }
			raftNode := gossipNode.GetSystemRaft()

			dt.gossipMutex.Lock()
			dt.gossipNodes = append(dt.gossipNodes, gossipNode)
			dt.gossipMutex.Unlock()

			dt.raftMutex.Lock()
			dt.raftNodes = append(dt.raftNodes, raftNode)
			dt.raftMutex.Unlock()

			wg.Done()
		}(params)
	}

	wg.Wait()

	dt.app.Logger.Infof("Started %d nodes", dt.nodeCount)

	ConcurrentTestCounter++
}

func (dt *TestCluster) Cleanup() {
	if ConcurrentTestCounter > 1 {
		ConcurrentTestCounter--
		return
	}
	if dt != nil {
		err := os.RemoveAll(dt.app.DataDir)
		if err != nil {
			panic(err)
		}
	}
	err := os.RemoveAll(TestDataDir)
	if err != nil {
		panic(err)
	}
}

func (dt *TestCluster) GetRaftLeaderNode() RaftNode {
	return dt.raftNodes[RaftLeaderID-1]
}

func (dt *TestCluster) GetRaftNode(index int) RaftNode {
	return dt.raftNodes[index]
}

func (dt *TestCluster) GetRaftNode1() RaftNode {
	return dt.raftNodes[0]
}

func (dt *TestCluster) GetRaftNode2() RaftNode {
	return dt.raftNodes[1]
}

func (dt *TestCluster) GetRaftNode3() RaftNode {
	return dt.raftNodes[2]
}

func (dt *TestCluster) CreateServerCluster() error {
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewServerConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, ClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(ClusterID, join, sm.CreateServerConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(ClusterID)
	return nil
}

func (dt *TestCluster) CreateOrganizationCluster() error {
	dt.app.Logger.Debugf("Creating organization cluster: %d", OrganizationClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewOrganizationConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, OrganizationClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(OrganizationClusterID, join, sm.CreateOrganizationConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(OrganizationClusterID)
	return nil
}

func (dt *TestCluster) CreateFarmConfigCluster(farmID uint64) error {
	dt.app.Logger.Debugf("Creating FarmConfig cluster: %d", farmID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		farmConfigChangeChan := make(chan config.Farm, 5)
		sm := statemachine.NewFarmConfigOnDiskStateMachine(dt.app.Logger, dt.app.IdGenerator,
			dt.app.DataDir, farmID, nodeID, farmConfigChangeChan)
		err := raftNode.CreateOnDiskCluster(farmID, join, sm.CreateFarmConfigOnDiskStateMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(farmID)
	return nil
}

func (dt *TestCluster) CreateFarmStateCluster(farmStateID uint64) error {
	dt.app.Logger.Debugf("Creating FarmState cluster: %d", farmStateID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		join := false
		farmStateChangeChan := make(chan state.FarmStateMap, 5)
		sm := statemachine.NewFarmStateConcurrentStateMachine(dt.app.Logger,
			farmStateID, farmStateChangeChan)
		err := raftNode.CreateConcurrentCluster(farmStateID, join, sm.CreateFarmStateConcurrentStateMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(farmStateID)
	return nil
}

func (dt *TestCluster) CreateDeviceConfigCluster(deviceID uint64) error {
	dt.app.Logger.Debugf("Creating device config cluster: %d", deviceID)
	for i := 0; i < dt.nodeCount; i++ {
		nodeID := uint64(i + 1)
		raftNode := dt.GetRaftNode(i)
		join := false
		sm := statemachine.NewDeviceConfigOnDiskStateMachine(dt.app.Logger, dt.app.IdGenerator,
			dt.app.DataDir, deviceID, nodeID)
		err := raftNode.CreateOnDiskCluster(deviceID, join, sm.CreateDeviceConfigOnDiskStateMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(deviceID)
	return nil
}

func (dt *TestCluster) CreateDeviceDataCluster(deviceID uint64) error {
	deviceDataClusterID := dt.app.IdGenerator.CreateDeviceDataClusterID(deviceID)
	dt.app.Logger.Debugf("Creating device data cluster: %d", deviceDataClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		nodeID := uint64(i + 1)
		raftNode := dt.GetRaftNode(i)
		join := false
		sm := statemachine.NewDeviceDataOnDiskStateMachine(dt.app.Logger, dt.app.IdGenerator,
			dt.app.DataDir, deviceDataClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(deviceDataClusterID, join, sm.CreateDeviceDataOnDiskStateMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(deviceDataClusterID)
	return nil
}

func (dt *TestCluster) CreateEventLogCluster(farmID uint64) error {
	eventLogClusterID := dt.app.IdGenerator.CreateEventLogClusterID(farmID)
	dt.app.Logger.Debugf("Creating event log cluster: %d", eventLogClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		nodeID := uint64(i + 1)
		raftNode := dt.GetRaftNode(i)
		join := false
		sm := statemachine.NewEventLogOnDiskStateMachine(dt.app.Logger, dt.app.IdGenerator,
			dt.app.DataDir, eventLogClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(eventLogClusterID, join, sm.CreateEventLogOnDiskStateMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(eventLogClusterID)
	return nil
}

// func (dt *TestCluster) CreatePermissionCluster() error {
// 	for i := 0; i < dt.nodeCount; i++ {
// 		raftNode := dt.GetRaftNode(i)
// 		nodeID := raftNode.GetParams().GetNodeID()
// 		join := false
// 		sm := statemachine.NewPermissionConfigMachine(dt.app.Logger,
// 			dt.app.IdGenerator, dt.app.DataDir, PermissionClusterID, nodeID)
// 		err := raftNode.CreateOnDiskCluster(PermissionClusterID, join, sm.CreatePermissionConfigMachine)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	dt.GetRaftNode(0).WaitForClusterReady(PermissionClusterID)
// 	return nil
// }

func (dt *TestCluster) CreateUserCluster() error {
	dt.app.Logger.Debugf("Creating user cluster: %d", UserClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewUserConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, UserClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(UserClusterID, join, sm.CreateUserConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(UserClusterID)
	return nil
}

func (dt *TestCluster) CreateRoleCluster() error {
	dt.app.Logger.Debugf("Creating role cluster: %d", RoleClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewRoleConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, RoleClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(RoleClusterID, join, sm.CreateRoleConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(RoleClusterID)
	return nil
}

func (dt *TestCluster) CreateAlgorithmCluster() error {
	dt.app.Logger.Debugf("Creating algoritm cluster: %d", AlgorithmClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewAlgorithmConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, AlgorithmClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(AlgorithmClusterID, join, sm.CreateAlgorithmConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(AlgorithmClusterID)
	return nil
}

func (dt *TestCluster) CreateRegistrationCluster() error {
	dt.app.Logger.Debugf("Creating registration cluster: %d", RegistrationClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewRegistrationConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, RegistrationClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(RegistrationClusterID, join, sm.CreateRegistrationConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(RegistrationClusterID)
	return nil
}

func (dt *TestCluster) CreateCustomerCluster() error {
	dt.app.Logger.Debugf("Creating customer cluster: %d", CustomerClusterID)
	for i := 0; i < dt.nodeCount; i++ {
		raftNode := dt.GetRaftNode(i)
		nodeID := raftNode.GetParams().GetNodeID()
		join := false
		sm := statemachine.NewCustomerConfigMachine(dt.app.Logger,
			dt.app.IdGenerator, dt.app.DataDir, CustomerClusterID, nodeID)
		err := raftNode.CreateOnDiskCluster(CustomerClusterID, join, sm.CreateCustomerConfigMachine)
		if err != nil {
			return err
		}
	}
	dt.GetRaftNode(0).WaitForClusterReady(CustomerClusterID)
	return nil
}

func (dt *TestCluster) StopClusters(farmIDs []uint64) {
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(ClusterID)
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(OrganizationClusterID)
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(FarmConfigClusterID)
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(FarmStateClusterID)
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(UserClusterID)
	dt.GetRaftLeaderNode().GetNodeHost().StopCluster(RoleClusterID)
	for _, farmID := range farmIDs {
		dt.GetRaftLeaderNode().GetNodeHost().StopCluster(farmID)
		dt.GetRaftLeaderNode().GetNodeHost().StopCluster(farmID)
	}
}

func (dt *TestCluster) createPeers(localIP string, startPort int) []string {
	peers := make([]string, dt.nodeCount)
	for i := 0; i < dt.nodeCount; i++ {
		peers[i] = fmt.Sprintf("%s:%d", localIP, startPort+i)
	}
	return peers
}
