//go:build cluster
// +build cluster

package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	logging "github.com/op/go-logging"
)

type RaftRestService interface {
	RequestLeaderTransfer(w http.ResponseWriter, r *http.Request)
}

type RaftRestServiceImpl struct {
	logger     *logging.Logger
	raftNode   cluster.RaftNode
	httpWriter response.HttpWriter
}

func NewRaftRestService(
	logger *logging.Logger,
	raftNode cluster.RaftNode,
	httpWriter response.HttpWriter) RaftRestService {

	return &RaftRestServiceImpl{
		logger:     logger,
		raftNode:   raftNode,
		httpWriter: httpWriter}
}

func (restService *RaftRestServiceImpl) RequestLeaderTransfer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["clusterID"], 10, 64)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	nodeID, err := strconv.ParseUint(params["nodeID"], 10, 64)
	if err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error400(w, r, err)
		return
	}
	if err := restService.raftNode.RequestLeaderTransfer(farmID, nodeID); err != nil {
		restService.logger.Error(err)
		restService.httpWriter.Error500(w, r, err)
		return
	}
	restService.httpWriter.Success200(w, r, farmID)
}
