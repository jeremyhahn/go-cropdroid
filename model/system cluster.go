//go:build cluster
// +build cluster

package model

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/cluster/util"
)

type ClusterSystemStruct struct {
	ClusterID               int             `json:"cluster_id"`
	NodeID                  uint64          `json:"node_id"`
	Mode                    string          `json:"mode"`
	Version                 *app.AppVersion `json:"version"`
	Farms                   int             `json:"farms"`
	Changefeeds             int             `json:"changefeeds"`
	NotificationQueueLength int             `json:"notificationQueueLength"`
	DeviceIndexLength       int             `json:"deviceIndexLength"`
	ChannelIndexLength      int             `json:"channelIndexLength"`
	Runtime                 *SystemRuntime  `json:"runtime"`
	RaftStats               *RaftStats      `json:"raft"`
	GossipStats             *GossipStats    `json:"gossip"`
}

type RaftStats struct {
	NumClusters int                      `json:"numClusters"`
	NumNodes    int                      `json:"nodes"`
	LeaderID    int                      `json:"leaderId"`
	IsLeader    bool                     `json:"leader"`
	IsReady     bool                     `json:"ready"`
	Params      *util.ClusterParams      `json:"params"`
	Clusters    []*cluster.ClusterStatus `json:"clusters"`
	Hashring    *Hashring                `json:"hashring"`
}

type GossipStats struct {
	NumClusters int               `json:"clusters"`
	NumNodes    int               `json:"nodes"`
	Datacenters int               `json:"datacenters"`
	Regions     int               `json:"regions"`
	HealthScore int               `json:"score"`
	Hashring    *Hashring         `json:"hashring"`
	Serf        map[string]string `json:"serf"`
}

type Hashring struct {
	Loads map[string]int64 `json:"loads"`
}
