package raft

import (
	"fmt"
	"os"
	"sync"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"

	clusterutil "github.com/jeremyhahn/go-cropdroid/cluster/util"
)

type LocalCluster struct {
	app           *app.App
	nodeCount     int
	gossipMutex   sync.Mutex
	gossipNodes   []cluster.GossipNode
	raftMutex     sync.Mutex
	raftNodes     []cluster.RaftNode
	clusterID     uint64
	userClusterID uint64
	roleClusterID uint64
	raftLeaderID  uint64
	dataDir       string
}

func NewLocalCluster(app *app.App, nodeCount int, clusterID uint64) *LocalCluster {
	if nodeCount < 3 {
		panic("NodeCount must be greater than 3")
	}
	if nodeCount > 9 {
		panic("NodeCount must be less than 9")
	}
	return &LocalCluster{
		app:           app,
		nodeCount:     nodeCount,
		clusterID:     clusterID,
		userClusterID: uint64(104),
		roleClusterID: uint64(105),
		raftLeaderID:  3}
}

func (localCluster *LocalCluster) StartCluster() {

	clusterIaasProvider := ""
	clusterRegion := ""
	clusterZone := ""
	localIP := util.ParseLocalIP()
	gossipStartPort := 60010
	raftStartPort := 60020

	gossipPeers := localCluster.createPeers(localIP, gossipStartPort)
	raftPeers := localCluster.createPeers(localIP, raftStartPort)

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

	for i := 0; i < localCluster.nodeCount; i++ {

		nodeID := uint64(i)
		clusterGossipPort := gossipStartPort + i
		clusterRaftPort := raftStartPort + i

		raftOptions := clusterutil.RaftOptions{
			Port:              clusterRaftPort,
			RequestedLeaderID: localCluster.raftLeaderID,
			SystemClusterID:   localCluster.clusterID,
			//PermissionClusterID: PermissionClusterID,
			UserClusterID: localCluster.userClusterID,
			RoleClusterID: localCluster.roleClusterID}

		params := clusterutil.NewClusterParams(localCluster.app.Logger, raftOptions, nodeID,
			clusterIaasProvider, clusterRegion, clusterZone, localCluster.app.DataDir, localAddress,
			localIP, gossipPeers, raftPeers, clusterJoin, clusterGossipPort,
			clusterRaftPort, localCluster.raftLeaderID, clusterVirtualNodes, clusterMaxNodes,
			clusterBootstrap, clusterInit, localCluster.app.IdGenerator, localCluster.app.IdSetter,
			farmProvisionerChan, farmDeprovisionerChan, farmTickerProvisionerChan)

		wg.Add(1)
		go func(params *clusterutil.ClusterParams) {

			gossipNode := cluster.NewGossipNode(params, clusterutil.NewHashring(clusterVirtualNodes))
			gossipNode.Join()
			go gossipNode.Run()

			// Need to pass in a new cluster ID
			// raftNode := cluster.NewRaftNode(params, util.NewHashring(1))
			// for raftNode == nil {
			// 	localCluster.app.Logger.Info("Waiting for enough nodes to build the Raft quorum...")
			// 	time.Sleep(1 * time.Second)
			// 	raftNode = gossipNode.GetSystemRaft()
			// }
			raftNode := gossipNode.GetSystemRaft()

			localCluster.gossipMutex.Lock()
			localCluster.gossipNodes = append(localCluster.gossipNodes, gossipNode)
			localCluster.gossipMutex.Unlock()

			localCluster.raftMutex.Lock()
			localCluster.raftNodes = append(localCluster.raftNodes, raftNode)
			localCluster.raftMutex.Unlock()

			wg.Done()
		}(params)
	}

	wg.Wait()

	localCluster.app.Logger.Infof("Started %d nodes", localCluster.nodeCount)
}

func (localCluster *LocalCluster) DestroyData() {
	if localCluster.app != nil {
		err := os.RemoveAll(localCluster.app.DataDir)
		if err != nil {
			panic(err)
		}
	}
	// err := os.RemoveAll(localCluster.dataDir)
	// if err != nil {
	// 	panic(err)
	// }
}

func (localCluster *LocalCluster) GetRaftLeaderNode() cluster.RaftNode {
	return localCluster.raftNodes[localCluster.raftLeaderID-1]
}

func (localCluster *LocalCluster) GetRaftNode(index int) cluster.RaftNode {
	return localCluster.raftNodes[index]
}

func (localCluster *LocalCluster) GetRaftNode1() cluster.RaftNode {
	return localCluster.raftNodes[0]
}

func (localCluster *LocalCluster) GetRaftNode2() cluster.RaftNode {
	return localCluster.raftNodes[1]
}

func (localCluster *LocalCluster) GetRaftNode3() cluster.RaftNode {
	return localCluster.raftNodes[2]
}

func (localCluster *LocalCluster) NodeCount() int {
	return len(localCluster.raftNodes)
}

func (localCluster *LocalCluster) StopCluster() {
	localCluster.GetRaftLeaderNode().GetNodeHost().StopCluster(localCluster.clusterID)
}

func (localCluster *LocalCluster) createPeers(localIP string, startPort int) []string {
	peers := make([]string, localCluster.nodeCount)
	for i := 0; i < localCluster.nodeCount; i++ {
		peers[i] = fmt.Sprintf("%s:%d", localIP, startPort+i)
	}
	return peers
}
