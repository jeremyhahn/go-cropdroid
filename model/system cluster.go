// +build cluster

package model

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
)

// https://github.com/google/pprof
// https://golang.org/pkg/net/http/pprof/
// https://golang.org/pkg/runtime

type SystemRuntime struct {
	Version     string `json:"version"`
	Goroutines  int    `json:"goroutines"`
	Cpus        int    `json:"cpus"`
	Cgo         int64  `json:"cgo"`
	HeapSize    uint64 `json:"heapAlloc"`
	Alloc       uint64 `json:"alloc"`
	Sys         uint64 `json:"sys"`
	Mallocs     uint64 `json:"mallocs"`
	Frees       uint64 `json:"frees"`
	NumGC       uint32 `json:"gc"`
	NumForcedGC uint32 `json:"forcedgc"`
}

type RaftStats struct {
	NumClusters int                      `json:"numClusters"`
	NumNodes    int                      `json:"nodes"`
	LeaderID    int                      `json:"leaderId"`
	IsLeader    bool                     `json:"leader"`
	IsReady     bool                     `json:"ready"`
	Params      *cluster.ClusterParams   `json:"params"`
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

type System struct {
	Mode                    string          `json:"mode"`
	Version                 *app.AppVersion `json:"version"`
	Farms                   int             `json:"farms"`
	Changefeeds             int             `json:"changefeeds"`
	NotificationQueueLength int             `json:"notificationQueueLength"`
	ControllerIndexLength   int             `json:"controllerIndexLength"`
	ChannelIndexLength      int             `json:"channelIndexLength"`
	Runtime                 *SystemRuntime  `json:"runtime"`
	RaftStats               *RaftStats      `json:"raft"`
	GossipStats             *GossipStats    `json:"gossip"`
}
