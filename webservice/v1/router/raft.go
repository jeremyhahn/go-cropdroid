//go:build cluster
// +build cluster

package router

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/op/go-logging"
)

type RaftRouter struct {
	middleware      middleware.JsonWebTokenMiddleware
	raftRestService rest.RaftRestService
	WebServiceRouter
}

// Creates a new web service raft router
func NewRaftRouter(
	logger *logging.Logger,
	raftNode cluster.RaftNode,
	middleware middleware.JsonWebTokenMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &RaftRouter{
		middleware: middleware,
		raftRestService: rest.NewRaftRestService(
			logger,
			raftNode,
			httpWriter)}
}

// Registers all of the raft endpoints at the root of the webservice (/api/v1)
func (raftRouter *RaftRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		raftRouter.transferLeader(router, baseURI)}
}

// @Summary Request leader transfer
// @Description Transfer raft leader to another node
// @Tags Raft
// @Produce  json
// @Param   clusterID	path	integer	true	"string valid"
// @Param   nodeID		path	integer	true	"string valid"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /raft/transfer/{clusterID}/{nodeID} [get]
// @Security JWT
func (raftRouter *RaftRouter) transferLeader(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/raft/transfer/{clusterID}/{nodeID}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(raftRouter.middleware.Validate),
		negroni.Wrap(http.HandlerFunc(raftRouter.raftRestService.RequestLeaderTransfer)),
	))
	return endpoint
}
