//go:build cluster
// +build cluster

package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
)

type RaftOptions struct {
	Port                  int    `json:"raftPort"`
	RequestedLeaderID     int    `json:"raftRequestedLeaderID"`
	OrganizationClusterID uint64 `json:"organizationId"`
	SystemClusterID       uint64 `json:"systemId"`
	UserClusterID         uint64 `json:"userId"`
	RoleClusterID         uint64 `json:"roleId"`
	AlgorithmClusterID    uint64 `json:"algorithmId`
	RegistrationClusterID uint64 `json:"registrationId`
}

type ClusterParams struct {
	Logger                    *logging.Logger `json:"-"`
	IdGenerator               util.IdGenerator
	IdSetter                  util.IdSetter
	RaftOptions               RaftOptions
	NodeName                  string           `json:"name"`
	ClusterID                 uint64           `json:"clusterId`
	NodeID                    uint64           `json:"nodeId`
	Provider                  string           `json:"provider"`
	Region                    string           `json:"region"`
	Zone                      string           `json:"zone"`
	DataDir                   string           `json:"dataDir"`
	LocalAddress              string           `json:"dataDir"`
	Listen                    string           `json:"listen"`
	Join                      bool             `json:"join"`
	Initialize                bool             `json:"initialize"`
	GossipPeers               []string         `json:"gossipPeers"`
	GossipPort                int              `json:"gossipPort"`
	Raft                      []string         `json:"raft"`
	Vnodes                    int              `json:"vnodes"`
	MaxNodes                  int              `json:"maxNodes"`
	Bootstrap                 int              `json:"bootstrap"`
	FarmProvisionerChan       chan config.Farm `json:"-"`
	FarmDeprovisionerChan     chan config.Farm `json:"-"`
	FarmTickerProvisionerChan chan uint64      `json:"-"`
}

func NewClusterParams(logger *logging.Logger, raftOptions RaftOptions, nodeID uint64, provider, region,
	zone, dataDir, localAddress, listen string, gossipPeers []string, raft []string, join bool,
	gossipPort, raftPort, raftRequestedLeaderID int, vnodes, maxNodes, bootstrap int, initialize bool,
	idGenerator util.IdGenerator, idSetter util.IdSetter, farmProvisionerChan chan config.Farm,
	farmDeprovisionerChan chan config.Farm, farmTickerProvisionerChan chan uint64) *ClusterParams {

	var nodeName string
	hostname, _ := os.Hostname()

	if bootstrap > 0 {

		logger.Debugf("Bootstrapping new cluster with raft groups: %v", raftOptions)

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

		nodeName = fmt.Sprintf("%s-%d-%d", hostname, raftOptions.SystemClusterID, nodeID)
		logger.Debugf("Assigning nodeName %s with id %d", nodeName, nodeID)
		//}
	} else {

		logger.Debugf("Leaving nodeID assignment to Gossip service. Hostname is %s", hostname)

		nodeID = uint64(len(gossipPeers) + 1)
		nodeName = fmt.Sprintf("%s-%d-%d", hostname, raftOptions.SystemClusterID, nodeID)

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
		Logger:                    logger,
		IdGenerator:               idGenerator,
		IdSetter:                  idSetter,
		RaftOptions:               raftOptions,
		NodeName:                  nodeName,
		ClusterID:                 raftOptions.SystemClusterID,
		NodeID:                    nodeID,
		Provider:                  provider,
		Region:                    region,
		Zone:                      zone,
		DataDir:                   dataDir,
		Listen:                    listen,
		GossipPeers:               gossipPeers,
		Join:                      join,
		Initialize:                initialize,
		GossipPort:                gossipPort,
		Raft:                      raft,
		Vnodes:                    vnodes,
		MaxNodes:                  maxNodes,
		Bootstrap:                 bootstrap,
		FarmProvisionerChan:       farmProvisionerChan,
		FarmDeprovisionerChan:     farmDeprovisionerChan,
		FarmTickerProvisionerChan: farmTickerProvisionerChan}
}

func (cp *ClusterParams) SetDataDir(dir string) {
	cp.DataDir = dir
}

func (cp *ClusterParams) GetNodeName() string {
	return cp.NodeName
}

func (cp *ClusterParams) GetNodeID() uint64 {
	return cp.NodeID
}

func (cp *ClusterParams) SetClusterID(id uint64) {
	cp.ClusterID = id
}

func (cp *ClusterParams) GetClusterID() uint64 {
	return cp.ClusterID
}

func (cp *ClusterParams) GetFarmProvisionerChan() chan config.Farm {
	return cp.FarmProvisionerChan
}

func (cp *ClusterParams) GetFarmDeprovisionerChan() chan config.Farm {
	return cp.FarmDeprovisionerChan
}

func (cp *ClusterParams) GetFarmTickerProvisionerChan() chan uint64 {
	return cp.FarmTickerProvisionerChan
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
