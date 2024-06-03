//go:build cluster
// +build cluster

package webservice

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"

	v1 "github.com/jeremyhahn/go-cropdroid/webservice/v1"
)

var (
	ErrFarmNotFound = errors.New("Farm not found")
)

type ClusterWebServerV1 struct {
	gossipNode cluster.GossipNode
	raftNode   cluster.RaftNode
	*WebServerV1
}

func NewClusterWebServerV1(
	app *app.App,
	gossipNode cluster.GossipNode,
	raftNode cluster.RaftNode,
	mapperRegistry mapper.MapperRegistry,
	serviceRegistry service.ServiceRegistry,
	restServiceRegistry rest.RestServiceRegistry,
	farmTickerProvisionerChan chan uint64) *ClusterWebServerV1 {

	webserver := NewWebServerV1(
		app,
		mapperRegistry,
		serviceRegistry,
		restServiceRegistry,
		farmTickerProvisionerChan)

	webserver.systemEventLogService = serviceRegistry.GetEventLogService(raftNode.GetParams().ClusterID)

	return &ClusterWebServerV1{
		WebServerV1: webserver,
		gossipNode:  gossipNode,
		raftNode:    raftNode}
}

func (server ClusterWebServerV1) RunClusterProvisionerConsumer() {
	for {
		select {
		case farmID := <-server.WebServerV1.farmTickerProvisionerChan:
			server.WebServerV1.app.Logger.Warningf("[Webserver.RunClusterProvisionerConsumer] Received message for farmID %d", farmID)
			server.buildRoutes()
		}
	}
}

func (server *ClusterWebServerV1) Run() {

	fs := http.FileServer(http.Dir(common.HTTP_PUBLIC_HTML))
	server.router.PathPrefix("/").Handler(fs)
	http.Handle("/", server.httpServer.Handler)

	server.buildRoutes()

	if server.app.WebTlsPort > 0 {
		go server.startHttps()
	} else {
		go server.startHttp()
	}

	server.app.DropPrivileges()

	<-server.closeChan
}

func (server ClusterWebServerV1) buildRoutes() {
	muxRouter := mux.NewRouter().StrictSlash(true)
	httpWriter := response.NewResponseWriter(server.app.Logger, nil)
	endpointList := v1.NewClusterRouterV1(
		server.app,
		server.raftNode,
		server.WebServerV1.mapperRegistry,
		server.WebServerV1.serviceRegistry.(service.ClusterServiceRegistry),
		server.WebServerV1.restServiceRegistry,
		server.WebServerV1.farmWebSocketHandler,
		muxRouter,
		httpWriter).RegisterRoutes(muxRouter, server.WebServerV1.baseURI)
	server.WebServerV1.routerMutex.Lock()
	server.WebServerV1.router = muxRouter
	server.WebServerV1.endpointList = endpointList
	server.WebServerV1.httpServer.Handler = server.WebServerV1.router
	server.WebServerV1.routerMutex.Unlock()
}
