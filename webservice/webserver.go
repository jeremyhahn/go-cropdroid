// +build !cluster

package webservice

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
	"github.com/jeremyhahn/go-cropdroid/webservice/websocket"
)

var (
	ErrFarmNotFound = errors.New("Farm not found")
)

type Webserver struct {
	mutex                     sync.Mutex
	app                       *app.App
	httpServer                *http.Server
	router                    *mux.Router
	baseURI                   string
	registry                  service.ServiceRegistry
	restServices              []rest.RestService
	eventType                 string
	endpointList              []string
	closeChan                 chan bool
	farmTickerProvisionerChan chan uint64
	farmHubs                  map[uint64]*websocket.FarmHub
	farmHubMutex              sync.Mutex
	jsonWebTokenService       service.JsonWebTokenService
	notificationService       service.NotificationService
	notificationHubs          map[int]*websocket.NotificationHub
	notificationHubMutex      sync.Mutex
	farmTickerSubrouter       *mux.Router
	eventLogService           service.EventLogService
}

func NewWebserver(app *app.App, serviceRegistry service.ServiceRegistry,
	restServices []rest.RestService, farmTickerProvisionerChan chan uint64) *Webserver {

	webserver := &Webserver{
		mutex:                     sync.Mutex{},
		app:                       app,
		router:                    mux.NewRouter().StrictSlash(true),
		baseURI:                   "/api/v1",
		registry:                  serviceRegistry,
		restServices:              restServices,
		eventType:                 "WebServer",
		endpointList:              make([]string, 0),
		closeChan:                 make(chan bool, 1),
		farmTickerProvisionerChan: farmTickerProvisionerChan,
		farmHubs:                  make(map[uint64]*websocket.FarmHub),
		farmHubMutex:              sync.Mutex{},
		jsonWebTokenService:       serviceRegistry.GetJsonWebTokenService(),
		notificationService:       serviceRegistry.GetNotificationService(),
		notificationHubs:          make(map[int]*websocket.NotificationHub),
		notificationHubMutex:      sync.Mutex{},
		eventLogService:           serviceRegistry.GetEventLogService()}
	webserver.httpServer = &http.Server{
		ReadTimeout:  common.HTTP_SERVER_READ_TIMEOUT,
		WriteTimeout: common.HTTP_SERVER_WRITE_TIMEOUT,
		IdleTimeout:  common.HTTP_SERVER_IDLE_TIMEOUT,
		Handler:      webserver.router,
	}
	return webserver
}

func (server *Webserver) RunProvisionerConsumer() {
	for {
		select {
		case farmID := <-server.farmTickerProvisionerChan:
			server.app.Logger.Warningf("[Webserver.RunProvisionerConsumer] Received message for farmID %d", farmID)
			server.buildRoutes()
		}
	}
}

func (server *Webserver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.mutex.Lock()
	router := server.router
	server.mutex.Unlock()
	router.ServeHTTP(w, r)
}

func (server *Webserver) Run() {

	server.buildRoutes()

	// Static content web server
	fs := http.FileServer(http.Dir("public_html"))
	server.router.PathPrefix("/").Handler(fs)
	http.Handle("/", server.httpServer.Handler)

	sPort := fmt.Sprintf(":%d", server.app.WebPort)
	if server.app.SSLFlag {

		server.app.Logger.Debugf("Starting web services on TLS port %d", server.app.WebPort)
		server.eventLogService.Create(server.eventType, fmt.Sprintf("Starting web server on TLS port %d", server.app.WebPort))

		certfile := fmt.Sprintf("%s/cert.pem", server.app.KeyDir)
		keyfile := fmt.Sprintf("%s/key.pem", server.app.KeyDir)

		cert, err := tls.LoadX509KeyPair(certfile, keyfile)
		if err != nil {
			server.app.Logger.Fatalf("[Webserver] Unable to load TLS certificates", err)
		}
		var tlsconf tls.Config
		tlsconf.Certificates = make([]tls.Certificate, 1)
		tlsconf.Certificates[0] = cert

		server.httpServer.TLSConfig = &tlsconf

		listener, err := tls.Listen("tcp4", sPort, &tlsconf)
		if err != nil {
			log.Fatalln("Unable to bind to SSL port: ", err)
		}

		server.app.DropPrivileges()

		if server.app.RedirectHttpToHttps {
			server.app.Logger.Debugf("[Webserver] Redirecting HTTP to HTTPS")
			go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+sPort+r.URL.String(), http.StatusMovedPermanently)
			}))
		}

		//err = http.Serve(listener, server.router)
		err = server.httpServer.Serve(listener)
		if err != nil {
			server.app.Logger.Fatalf("[WebServer] Unable to start web server: %s", err.Error())
		}

	} else {

		server.app.Logger.Infof("Starting web services on port %d", server.app.WebPort)
		server.eventLogService.Create(server.eventType, fmt.Sprintf("Starting web services on port %d", server.app.WebPort))

		ipv4Listener, err := net.Listen("tcp4", sPort)
		if err != nil {
			log.Fatal(err)
		}

		server.app.DropPrivileges()

		//err = http.Serve(ipv4Listener, server.router)
		err = server.httpServer.Serve(ipv4Listener)
		if err != nil {
			server.app.Logger.Fatalf("[WebServer] Unable to start web server: %s", err.Error())
		}
	}
}

func (server *Webserver) buildRoutes() {

	router := mux.NewRouter().StrictSlash(true)
	endpointList := make([]string, 0)

	jsonWriter := rest.NewJsonWriter()
	//baseOrgURI := fmt.Sprintf("%s/organizations/{organizationID}", server.baseURI)
	baseFarmURI := fmt.Sprintf("%s/farms/{farmID}", server.baseURI)
	//baseFarmURI := fmt.Sprintf("%s/farms/{farmID}", baseOrgURI)

	registrationService := rest.NewRegisterRestService(server.app, server.registry.GetUserService(), jsonWriter)

	// REST Handlers - Public Access
	router.HandleFunc("/endpoints", server.endpoints)
	router.HandleFunc("/system", server.systemStatus)
	router.HandleFunc("/api/v1/pubkey", server.publicKey)
	router.HandleFunc("/api/v1/register", registrationService.Register)
	router.HandleFunc("/api/v1/login", server.jsonWebTokenService.GenerateToken)
	router.HandleFunc("/api/v1/login/refresh", server.jsonWebTokenService.RefreshToken)
	endpointList = append(endpointList, "/api/v1/register")
	endpointList = append(endpointList, "/api/v1/login")

	router.HandleFunc(fmt.Sprintf("%s/notification/{type}/{message}", baseFarmURI), server.sendNotification)
	router.HandleFunc(fmt.Sprintf("%s/notification/{type}/{message}/{priority}", baseFarmURI), server.sendNotification)

	/*
		farmStateURI := fmt.Sprintf("%s/state", server.baseURI)
		router.HandleFunc(farmStateURI, server.state)
		endpointList = append(endpointList, farmStateURI)

		farmConfigURI := fmt.Sprintf("%s/config", server.baseURI)
		router.HandleFunc(farmConfigURI, server.config)
		endpointList = append(endpointList, farmConfigURI)
	*/
	//router.HandleFunc("/maint/{mode}", server.MaintenanceMode).Methods("GET")
	//farmMaintModeURI := fmt.Sprintf("%s/maint/{mode}", baseFarmURI)
	//router.HandleFunc(farmMaintModeURI, server.MaintenanceMode).Methods("GET")
	//endpointList = append(endpointList, farmMaintModeURI)

	for _, restService := range server.restServices {
		endpoints := restService.RegisterEndpoints(router, server.baseURI, baseFarmURI)
		for _, endpoint := range endpoints {
			server.app.Logger.Debugf("[WebServer.buildRoutes] Loading REST resource: %s", endpoint)
			endpointList = append(endpointList, endpoint)
		}
	}

	endpoint := fmt.Sprintf("%s/events", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(server.jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(server.events)),
	))
	endpointList = append(endpointList, endpoint)

	endpoint = fmt.Sprintf("%s/events/{page}", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(server.jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(server.eventsPage)),
	))
	endpointList = append(endpointList, endpoint)

	// /virtual
	/*
		router.Handle("/api/v1/virtual/{vdevice}/{metric}/{value}", negroni.New(
			negroni.HandlerFunc(server.jsonWebTokenService.Validate),
			negroni.Wrap(http.HandlerFunc(server.setVirtualMetric)),
		)).Methods("GET")
	*/

	// Websocket hubs

	// /api/v1/farms/{farmID}/notifications
	endpoint = fmt.Sprintf("%s/notifications", baseFarmURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(server.jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := mux.Vars(r)
			farmID, err := strconv.Atoi(params["farmID"])
			if err != nil {
				rest.BadRequestError(w, r, err, jsonWriter)
				return
			}
			if _, ok := server.notificationHubs[farmID]; !ok {
				server.app.Logger.Debugf("[WebServer.buildRoutes] Creating new websocket notification hub for farm %d", farmID)
				server.notificationHubMutex.Lock()
				server.notificationHubs[farmID] = websocket.NewNotificationHub(server.app.Logger, server.notificationService)
				server.notificationHubMutex.Unlock()
				go server.notificationHubs[farmID].Run()
			}
			handler := websocket.NewNotificationHandler(server.app.Logger, server.notificationHubs[farmID], server.jsonWebTokenService)
			handler.OnConnect(w, r)
		})),
	))
	endpointList = append(endpointList, endpoint)

	//endpoint = fmt.Sprintf("%s/farmticker", server.farmTicker)
	//server.farmTickerSubrouter = router.PathPrefix(endpoint).Subrouter()
	//endpointList = append(endpointList, endpoint)

	// /api/v1/farmticker/{farmID}
	endpoint = fmt.Sprintf("%s/farmticker/{farmID}", server.baseURI)
	router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(server.jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := mux.Vars(r)
			farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
			if err != nil {
				rest.BadRequestError(w, r, err, jsonWriter)
				return
			}
			farmService := server.registry.GetFarmService(farmID)
			if farmService == nil {
				rest.BadRequestError(w, r, ErrFarmNotFound, jsonWriter)
				return
			}
			if _, ok := server.farmHubs[farmID]; !ok {
				server.app.Logger.Debugf("[WebServer.buildRoutes] Creating new websocket hub for farm %d", farmID)
				server.farmHubMutex.Lock()
				server.farmHubs[farmID] = websocket.NewFarmHub(server.app.Logger, server.notificationService, farmService)
				server.farmHubMutex.Unlock()
				go server.farmHubs[farmID].Run()
			}
			handler := websocket.NewFarmHandler(server.app.Logger, server.farmHubs[farmID], server.notificationService, server.jsonWebTokenService)
			handler.OnConnect(w, r)
		})),
	))
	endpointList = append(endpointList, endpoint)

	server.app.Logger.Debugf("[WebServer.buildRoutes] Loaded %d REST services", len(endpointList))

	sort.Strings(endpointList)

	server.mutex.Lock()
	server.router = router
	server.endpointList = endpointList
	server.httpServer.Handler = server.router
	server.mutex.Unlock()
}

/*
func (server *Webserver) Shutdown() {
	server.app.Logger.Info("Shutting down web services")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.httpServer.Shutdown(ctx)
	cancel()
}*/

func (server *Webserver) farmTicker(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	farmid := params["farmID"]
	farmID, err := strconv.ParseUint(farmid, 10, 64)
	if err != nil {
		server.app.Logger.Errorf("[Webserver.farmTicker] Error: %s", err)
		return
	}

	farmService := server.registry.GetFarmService(farmID)
	if farmService == nil {
		server.app.Logger.Error("[Webserver.farmTicker] Can find farm service with farmID %d", farmID)
		return
	}

	farmHub := websocket.NewFarmHub(server.app.Logger, server.notificationService, farmService)
	go farmHub.Run()

	endpoint := fmt.Sprintf("%s/farmticker/%d", server.baseURI, farmID)
	server.router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(server.jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := websocket.NewFarmHandler(server.app.Logger, farmHub, server.notificationService, server.jsonWebTokenService)
			handler.OnConnect(w, r)
		})),
	))
	server.endpointList = append(server.endpointList, endpoint)
}

func (server *Webserver) sendNotification(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	_type := params["type"]
	message := params["message"]

	priority, err := strconv.Atoi(params["priority"])
	if err != nil {
		priority = common.NOTIFICATION_PRIORITY_LOW
	}

	server.registry.GetNotificationService().Enqueue(&model.Notification{
		Device:    "webserver",
		Priority:  priority,
		Type:      _type,
		Message:   message,
		Timestamp: time.Now()})
}

func (server *Webserver) state(w http.ResponseWriter, r *http.Request) {
	states := make([]state.FarmStateMap, 0)
	for _, farmService := range server.registry.GetFarmServices() {
		states = append(states, farmService.GetState())
	}
	rest.NewJsonWriter().Write(w, http.StatusOK, states)
}

func (server *Webserver) config(w http.ResponseWriter, r *http.Request) {
	rest.NewJsonWriter().Write(w, http.StatusOK, server.app.Config)
}

func (server *Webserver) endpoints(w http.ResponseWriter, r *http.Request) {
	server.WalkRoutes() // TODO remove
	rest.NewJsonWriter().Write(w, http.StatusOK, server.endpointList)
}

func (server *Webserver) systemStatus(w http.ResponseWriter, r *http.Request) {
	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)
	changefeedCount := 0
	if changefeedService := server.registry.GetChangefeedService(); changefeedService != nil {
		changefeedCount = changefeedService.FeedCount()
	}
	systemStatus := &model.System{
		Mode:                    server.app.Mode,
		NotificationQueueLength: server.registry.GetNotificationService().QueueSize(),
		Farms:                   len(server.registry.GetFarmServices()),
		Changefeeds:             changefeedCount,
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

	rest.NewJsonWriter().Write(w, http.StatusOK, systemStatus)
}

func (server *Webserver) WalkRoutes() {
	err := server.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			server.app.Logger.Debug("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			server.app.Logger.Debug("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			server.app.Logger.Debug("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			server.app.Logger.Debug("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			server.app.Logger.Debug("Methods:", strings.Join(methods, ","))
		}
		server.app.Logger.Debug("")
		return nil
	})
	if err != nil {
		server.app.Logger.Errorf("[Webserver.WalkRoutes] Error: err", err)
	}
}

func (server *Webserver) MaintenanceMode(w http.ResponseWriter, r *http.Request) {

	//server.eventLogService.Create(server.eventType,
	//	fmt.Sprintf("/maint/%d requested by %s", mode, server.clientIP(r)))

	params := mux.Vars(r)

	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
	if err != nil {
		server.app.Logger.Error(err.Error())
		server.sendBadRequest(w, r, err)
		return
	}

	/*
		mode, err := strconv.Atoi(params["mode"])
		if err != nil {
			server.app.Logger.Error(err.Error())
			server.sendBadRequest(w, r, err)
			return
		}*/

	farmService := server.registry.GetFarmService(farmID)
	if farmService == nil {
		server.app.Logger.Error(ErrFarmNotFound)
		return
	}

	farmState := farmService.GetState()
	if farmState == nil {
		server.app.Logger.Error(err.Error())
		server.sendBadRequest(w, r, err)
		return
	}

	/*
		if mode == 0 {
			farmState.SetMaintenanceMode(false)
			server.app.FarmStore.Put(farmID, farmState)
		} else {
			farmState.SetMaintenanceMode(true)
			server.app.FarmStore.Put(farmID, farmState)
		}*/

	w.Header().Set("Content-Type", "application/json")
	json, _ := json.MarshalIndent(farmState, "", " ")
	fmt.Fprintln(w, string(json))
}

func (server *Webserver) publicKey(w http.ResponseWriter, r *http.Request) {
	pubkey := server.app.KeyPair.GetPublicBytes()

	//encoded := base64.StdEncoding.EncodeToString(pubkey)
	//rest.NewJsonWriter().Write(w, http.StatusOK, encoded)

	rest.NewJsonWriter().Write(w, http.StatusOK, string(pubkey))
}

func (server *Webserver) events(w http.ResponseWriter, r *http.Request) {
	//server.eventLogService.Create(server.eventType,
	//	fmt.Sprintf("/events requested by %s", server.clientIP(r)))

	entities := server.registry.GetEventLogService().GetAll()
	w.Header().Set("Content-Type", "application/json")
	json, _ := json.MarshalIndent(entities, "", " ")
	fmt.Fprintln(w, string(json))
}

func (server *Webserver) eventsPage(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	page := params["page"]

	//server.eventLogService.Create(server.eventType,
	//	fmt.Sprintf("/eventsPage/%s requested by %s", page, server.clientIP(r)))

	p, err := strconv.ParseInt(page, 10, 0)
	if err != nil {
		server.sendBadRequest(w, r, err)
	}

	server.app.Logger.Debugf("[Webserver.eventsPage] page %s requested", page)

	entities := server.registry.GetEventLogService().GetPage(p)
	w.Header().Set("Content-Type", "application/json")
	json, _ := json.MarshalIndent(entities, "", " ")
	fmt.Fprintln(w, string(json))
}

func (server *Webserver) clientIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		server.app.Logger.Error(err.Error())
	}
	return ip
}

func (server *Webserver) sendBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Bad Request", http.StatusBadRequest)
	fmt.Fprintln(w, err)
}

func (server *Webserver) sendInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	fmt.Fprintln(w, err)
}
