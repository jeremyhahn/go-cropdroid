//go:build cluster
// +build cluster

package rest

import (
	"net/http"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
)

type ClusterSystemRestService struct {
	gossipNode cluster.GossipNode
	raftNode   cluster.RaftNode
	SystemRestService
}

func NewClusterSystemRestService(
	app *app.App,
	serviceRegistry service.ClusterServiceRegistry,
	httpWriter response.HttpWriter,
	endpointList *[]string) SystemRestServicer {

	return &ClusterSystemRestService{
		gossipNode: serviceRegistry.GetGossipNode(),
		raftNode:   serviceRegistry.GetRaftNode(),
		SystemRestService: SystemRestService{
			app:             app,
			serviceRegistry: serviceRegistry,
			httpWriter:      httpWriter,
			endpointList:    endpointList}}
}

// Writes a page of system event log entries
func (restService *ClusterSystemRestService) EventsPage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	page := params["page"]
	p, err := strconv.Atoi(page)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	clusterID := restService.raftNode.GetParams().ClusterID
	pageQuery := query.NewPageQuery()
	pageQuery.Page = p
	pageQuery.SortOrder = query.SORT_DESCENDING
	pageResult, err := restService.serviceRegistry.GetEventLogService(clusterID).GetPage(pageQuery, common.CONSISTENCY_LOCAL)
	if err != nil {
		restService.httpWriter.Error400(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, pageResult)
}

// Writes the current system status and metrics
func (restService *ClusterSystemRestService) Status(w http.ResponseWriter, r *http.Request) {
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)
	systemStatus := &model.ClusterSystemStruct{
		ClusterID:               int(restService.app.ClusterID),
		NodeID:                  restService.app.NodeID,
		Mode:                    restService.app.Mode,
		NotificationQueueLength: restService.serviceRegistry.GetNotificationService().QueueSize(),
		Farms:                   len(restService.serviceRegistry.GetFarmServices()),
		//DeviceIndexLength:       server.app.DeviceIndex.Len(),
		//ChannelIndexLength:      server.app.ChannelIndex.Len(),
		Version: &app.AppVersion{
			Release:   app.Release,
			BuildDate: app.BuildDate,
			BuildUser: app.BuildUser,
			GitTag:    app.GitTag,
			GitHash:   app.GitHash,
			Image:     app.Image},
		Runtime: &model.SystemRuntime{
			Version:     runtime.Version(),
			Cpus:        runtime.NumCPU(),
			Cgo:         runtime.NumCgoCall(),
			Goroutines:  runtime.NumGoroutine(),
			HeapSize:    memstats.HeapAlloc, // essentially what the profiler is giving you (active heap memory)
			Alloc:       memstats.Alloc,     // similar to HeapAlloc, but for all go managed memory
			Sys:         memstats.Sys,       // the total amount of memory (address space) requested from the OS
			Mallocs:     memstats.Mallocs,
			Frees:       memstats.Frees,
			NumGC:       memstats.NumGC,
			NumForcedGC: memstats.NumForcedGC}}

	if restService.gossipNode != nil {
		systemStatus.GossipStats = &model.GossipStats{
			NumNodes: restService.gossipNode.GetMemberCount(),
			// HealthScore: server.gossipNode.GetHealthScore(),
			Serf:     restService.gossipNode.GetSerfStats(),
			Hashring: &model.Hashring{Loads: restService.gossipNode.GetHashring().GetLoads()}}
	}

	if restService.raftNode != nil {
		clusterID := restService.raftNode.GetParams().GetClusterID()
		leaderID, ready, _ := restService.raftNode.GetNodeHost().GetLeaderID(clusterID)
		systemStatus.RaftStats = &model.RaftStats{
			NumClusters: restService.raftNode.GetClusterCount(),
			NumNodes:    restService.raftNode.GetNodeCount(),
			LeaderID:    int(leaderID),
			IsLeader:    restService.raftNode.IsLeader(clusterID),
			IsReady:     ready,
			Params:      restService.raftNode.GetParams(),
			Clusters:    restService.raftNode.GetClusterStatus(),
			Hashring:    &model.Hashring{Loads: restService.raftNode.GetHashring().GetLoads()}}
	}

	restService.httpWriter.Success200(w, r, systemStatus)
}
