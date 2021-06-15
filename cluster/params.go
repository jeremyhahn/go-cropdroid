// +build cluster

package cluster

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
)

type ClusterParams struct {
	logger                    *logging.Logger             `json:"-"`
	nodeName                  string                      `json:name`
	clusterID                 uint64                      `json:"clusterId`
	nodeID                    uint64                      `json:"nodeId`
	provider                  string                      `json:"provider"`
	region                    string                      `json:"region"`
	zone                      string                      `json:"zone"`
	dataDir                   string                      `json:"dataDir"`
	localAddress              string                      `json:"dataDir"`
	listen                    string                      `json:"listen"`
	join                      bool                        `json:"join"`
	gossipPeers               []string                    `json:"gossipPeers"`
	gossipPort                int                         `json:"gossipPort"`
	raft                      []string                    `json:"raft"`
	raftPort                  int                         `json:"raftPort"`
	raftRequestedLeaderID     int                         `json:"raftRequestedLeaderID"`
	vnodes                    int                         `json:"vnodes"`
	maxNodes                  int                         `json:"maxNodes"`
	bootstrap                 int                         `json:"bootstrap"`
	farmProvisionerChan       chan config.FarmConfig      `json:"-"`
	farmTickerProvisionerChan chan uint64                 `json:"-"`
	daoRegistry               datastore.DatastoreRegistry `json:"-"`
}

func NewClusterParams(logger *logging.Logger, clusterID, nodeID uint64, provider, region,
	zone, dataDir, localAddress, listen string, gossipPeers []string, raft []string, join bool,
	gossipPort, raftPort, raftRequestedLeaderID int, vnodes, maxNodes, bootstrap int,
	daoRegistry datastore.DatastoreRegistry, farmProvisionerChan chan config.FarmConfig,
	farmTickerProvisionerChan chan uint64) *ClusterParams {

	var nodeName string
	hostname, _ := os.Hostname()

	if bootstrap > 0 {

		logger.Debugf("Bootstrapping new cluster id: %d", clusterID)

		// if gossipPeers[0] == "" {
		// 	gossipPeers = make([]string, 0)
		// 	nodeID = uint64(0)
		// 	nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)
		// 	logger.Debugf("Assigning initial bootstrap node %s id %d", nodeName, nodeID)
		// } else {

		for _nodeID, peer := range raft {

			pieces := strings.Split(peer, ":")
			parsedAddress := pieces[0]
			parsedPort := pieces[1]
			intParsedPort, err := strconv.ParseInt(parsedPort, 10, 0)
			if err != nil {
				logger.Fatal(err)
			}

			namePieces := strings.Split(pieces[0], ".")
			parsedHostname := namePieces[0]

			logger.Debugf("Comparing parsed local address %s:%s with actual local address %s and configured raft port %d",
				pieces[0], parsedPort, localAddress, raftPort)

			logger.Debugf("Comparing parsed name %s with hostname %s", parsedHostname, hostname)

			if (parsedAddress == listen && parsedAddress == localAddress &&
				int(intParsedPort) == raftPort) || parsedHostname == hostname {
				// Found the host in the array of raft peers, use its
				// ordinal position to assign a node id to this host

				//nodeID = uint64(_nodeID + 1)
				nodeID = uint64(_nodeID)
				logger.Debugf("Assigning member node id %d", nodeID)
				break
			}
		}

		nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)
		logger.Debugf("Assigning nodeName %s with id %d", nodeName, nodeID)
		//}
	} else {

		logger.Debugf("Leaving nodeID assignment to Gossip service. Hostname is %s", hostname)

		nodeID = uint64(len(gossipPeers) + 1)
		nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)

		logger.Debugf("Assigning nodeName %s", nodeName)
	}

	// This logic needs to be moved to gossip when the node joins the cluster
	// so the current cluster member count can be calculated.
	//
	// else {
	// 	nodeID = uint64(len(gossipPeers) + 1)
	// 	nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)
	// }

	return &ClusterParams{
		logger:                    logger,
		nodeName:                  nodeName,
		clusterID:                 clusterID,
		nodeID:                    nodeID,
		provider:                  provider,
		region:                    region,
		zone:                      zone,
		dataDir:                   dataDir,
		listen:                    listen,
		gossipPeers:               gossipPeers,
		join:                      join,
		gossipPort:                gossipPort,
		raft:                      raft,
		raftPort:                  raftPort,
		raftRequestedLeaderID:     raftRequestedLeaderID,
		vnodes:                    vnodes,
		maxNodes:                  maxNodes,
		bootstrap:                 bootstrap,
		farmProvisionerChan:       farmProvisionerChan,
		farmTickerProvisionerChan: farmTickerProvisionerChan,
		daoRegistry:               daoRegistry}
}

func (cp *ClusterParams) SetDataDir(dir string) {
	cp.dataDir = dir
}

func (cp *ClusterParams) GetNodeName() string {
	return cp.nodeName
}

func (cp *ClusterParams) GetNodeID() uint64 {
	return cp.nodeID
}

func (cp *ClusterParams) SetClusterID(id uint64) {
	cp.clusterID = id
}

func (cp *ClusterParams) GetClusterID() uint64 {
	return cp.clusterID
}

func (cp *ClusterParams) GetDatastoreRegistry() datastore.DatastoreRegistry {
	return cp.daoRegistry
}

func (cp *ClusterParams) GetFarmProvisionerChan() chan config.FarmConfig {
	return cp.farmProvisionerChan
}

func (cp *ClusterParams) GetFarmTickerProvisionerChan() chan uint64 {
	return cp.farmTickerProvisionerChan
}

/*
func (cp *ClusterParams) GetRegion() string {
	return cp.region
}

func (cp *ClusterParams) GetZone() string {
	return cp.zone
}

func (cp *ClusterParams) GetDataDir() string {
	return cp.dataDir
}

func (cp *ClusterParams) GetMaxNodes() int {
	return cp.maxNodes
}

func (cp *ClusterParams) GetVirtualNodes() int {
	return cp.vnodes
}

func (cp *ClusterParams) GetLogger() *logging.Logger {
	return cp.logger
}
*/
