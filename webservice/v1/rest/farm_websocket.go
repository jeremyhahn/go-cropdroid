package rest

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/websocket"
	logging "github.com/op/go-logging"
)

type FarmWebSocketRestServicer interface {
	FarmTickerConnect(w http.ResponseWriter, r *http.Request)
	PushNotificationConnect(w http.ResponseWriter, r *http.Request)
}

type FarmWebSocketRestService struct {
	logger                *logging.Logger
	farmHubs              map[uint64]*websocket.FarmHub
	farmHubsMutex         *sync.RWMutex
	notificationHubs      map[uint64]*websocket.NotificationHub
	notificationHubsMutex *sync.RWMutex
	serviceRegistry       service.ServiceRegistry
	middleware            middleware.JsonWebTokenMiddleware
	responseWriter        response.HttpWriter
	FarmWebSocketRestServicer
}

func NewFarmWebSocketRestService(
	logger *logging.Logger,
	farmHubs map[uint64]*websocket.FarmHub,
	farmHubsMutex *sync.RWMutex,
	notificationHubs map[uint64]*websocket.NotificationHub,
	notificationHubsMutex *sync.RWMutex,
	serviceRegistry service.ServiceRegistry,
	middleware middleware.JsonWebTokenMiddleware,
	responseWriter response.HttpWriter) FarmWebSocketRestServicer {

	return &FarmWebSocketRestService{
		logger:                logger,
		farmHubs:              farmHubs,
		farmHubsMutex:         farmHubsMutex,
		notificationHubs:      notificationHubs,
		notificationHubsMutex: notificationHubsMutex,
		serviceRegistry:       serviceRegistry,
		middleware:            middleware,
		responseWriter:        responseWriter}
}

// Handles a new connection request on the standard HTTP protocol for real-time updates to
// the farm state via websocket. This method creates a new FarmService using the passed HTTP
// GET farmID parameter, creates a new ticker hub if needed, upgrades the HTTP connection
// to a websocket, and connects the incoming request to the hub to start receiving updates.
func (farmHandler *FarmWebSocketRestService) FarmTickerConnect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		farmHandler.responseWriter.Error400(w, r, err)
		return
	}
	// Check to see if the ticker hub already exists using a fast read preferred lock
	farmHandler.farmHubsMutex.RLock()
	farmHub, exists := farmHandler.farmHubs[farmID]
	farmHandler.farmHubsMutex.RUnlock()
	if !exists {
		// Look up the FarmService
		farmService := farmHandler.serviceRegistry.GetFarmService(farmID)
		if farmService == nil {
			farmHandler.responseWriter.Error400(w, r, response.ErrFarmNotFound)
			return
		}
		farmHandler.logger.Debugf("Creating new websocket hub for farm %d", farmID)
		// Create and run a new farm ticker hub
		farmHandler.farmHubsMutex.Lock()
		farmHandler.farmHubs[farmID] = websocket.NewFarmHub(farmHandler.logger, farmService)
		farmHub = farmHandler.farmHubs[farmID]
		farmHandler.farmHubsMutex.Unlock()
		go farmHandler.farmHubs[farmID].Run()
	}
	// Upgrade the standard HTTP connection to a websocket
	websocket.NewFarmWebSocket(
		farmHandler.logger,
		farmHub,
		farmHandler.middleware,
		farmHandler.responseWriter).OnConnect(w, r)
}

// Handles a new connection request on the standard HTTP protocol for real-time push
// notifications via websocket. This method creates a new NotificationService using the
// passed HTTP GET farmID parameter, creates a new notification hub if needed, upgrades
// the HTTP connection to a websocket, and connects the incoming request to the hub to
// start receiving real-time push notifications.
func (farmHandler *FarmWebSocketRestService) PushNotificationConnect(w http.ResponseWriter, r *http.Request) {
	notificationService := farmHandler.serviceRegistry.GetNotificationService()
	params := mux.Vars(r)
	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		farmHandler.responseWriter.Error400(w, r, err)
		return
	}
	// Check to see if the notification hub already exists using a fast read preferred lock
	farmHandler.notificationHubsMutex.RLock()
	_, exists := farmHandler.notificationHubs[farmID]
	farmHandler.notificationHubsMutex.RUnlock()
	if !exists {
		// Create and run a new farm notification hub
		farmHandler.logger.Debugf("[WebServer.buildRoutes] Creating new websocket notification hub for farm %d", farmID)
		farmHandler.notificationHubsMutex.Lock()
		farmHandler.notificationHubs[farmID] = websocket.NewNotificationHub(farmHandler.logger, notificationService)
		farmHandler.notificationHubsMutex.Unlock()
		go farmHandler.notificationHubs[farmID].Run()
	}
	handler := websocket.NewNotificationWebSocket(farmHandler.logger,
		farmHandler.notificationHubs[farmID], farmHandler.middleware)
	handler.OnConnect(w, r)
}
