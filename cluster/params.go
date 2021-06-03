// +build cluster

package cluster

import (
	"fmt"
	"os"
	"strings"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore"
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
	listen                    string                      `json:"listen"`
	join                      bool                        `json:"join"`
	peers                     []string                    `json:"peers"`
	gossipPort                int                         `json:"gossipPort"`
	raft                      []string                    `json:"raft"`
	raftPort                  int                         `json:"raftPort"`
	vnodes                    int                         `json:"vnodes"`
	maxNodes                  int                         `json:"maxNodes"`
	bootstrap                 int                         `json:"bootstrap"`
	farmProvisionerChan       chan config.FarmConfig      `json:"-"`
	farmTickerProvisionerChan chan int                    `json:"-"`
	daoRegistry               datastore.DatastoreRegistry `json:"-"`
}

func NewClusterParams(logger *logging.Logger, clusterID, nodeID uint64, provider, region, zone, dataDir, listen string,
	peers []string, raft []string, join bool, gossipPort, raftPort int, vnodes, maxNodes, bootstrap int,
	daoRegistry datastore.DatastoreRegistry, farmProvisionerChan chan config.FarmConfig,
	farmTickerProvisionerChan chan int) *ClusterParams {

	var nodeName string
	hostname, _ := os.Hostname()

	if bootstrap > 0 {
		if peers[0] == "" {
			peers = make([]string, 0)
			nodeID = uint64(0)

			logger.Debugf("Assigning nodeID %d", nodeID)

			nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)
		} else {
			for _nodeID, peer := range peers {
				pieces := strings.Split(peer, ":")
				namePieces := strings.Split(pieces[0], ".")
				if namePieces[0] == hostname {

					nodeID = uint64(_nodeID + 1)

					logger.Debugf("Assigning nodeID %d", nodeID)
				}
			}
			nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)
		}
	} else {
		logger.Debugf("Leaving nodeID assignment up to Gossip service. Hostname is %s", hostname)

		nodeID = uint64(len(peers) + 1)
		nodeName = fmt.Sprintf("%s-%d-%d", hostname, clusterID, nodeID)

		logger.Debugf("Assigning nodeName %s", nodeName)
	}

	//
	// This logic needs to be moved to gossip when the node joins the cluster
	//
	// else {
	// 	nodeID = uint64(len(peers) + 1)
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
		peers:                     peers,
		join:                      join,
		gossipPort:                gossipPort,
		raft:                      raft,
		raftPort:                  raftPort,
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

func (cp *ClusterParams) GetFarmTickerProvisionerChan() chan int {
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
