// +build webserver_original

package webservice

import (
	"crypto/tls"
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"sort"
	"strconv"
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

type Webserver struct {
	app          *app.App
	router       *mux.Router
	httpServer   *http.Server
	registry     service.ServiceRegistry
	restServices []rest.RestService
	eventType    string
	endpointList []string
	closeChan    chan bool
}

func NewWebserver(app *app.App, serviceRegistry service.ServiceRegistry, restServices []rest.RestService) *Webserver {
	router := mux.NewRouter().StrictSlash(true)
	return &Webserver{
		app:    app,
		router: router,
		httpServer: &http.Server{
			ReadTimeout:  common.HTTP_SERVER_READ_TIMEOUT,
			WriteTimeout: common.HTTP_SERVER_WRITE_TIMEOUT,
			IdleTimeout:  common.HTTP_SERVER_IDLE_TIMEOUT,
			Handler:      router,
		},
		registry:     serviceRegistry,
		restServices: restServices,
		eventType:    "WebServer",
		endpointList: make([]string, 0),
		closeChan:    make(chan bool, 1)}
}

func (server *Webserver) Run() {

	jsonWriter := rest.NewJsonWriter()
	sPort := fmt.Sprintf(":%d", server.app.WebPort)
	baseURI := "/api/v1"
	//baseOrgURI := fmt.Sprintf("%s/organizations/{organizationID}", baseURI)
	baseFarmURI := fmt.Sprintf("%s/farms/{farmID}", baseURI)
	//baseFarmURI := fmt.Sprintf("%s/farms/{farmID}", baseOrgURI)

	registrationService := rest.NewRegisterRestService(server.app, server.registry.GetUserService(), jsonWriter)
	jsonWebTokenService := server.registry.GetJsonWebTokenService()
	notificationService := server.registry.GetNotificationService()
	eventLogService := server.registry.GetEventLogService()

	// Static content web server
	fs := http.FileServer(http.Dir("public_html"))

	// REST Handlers - Public Access
	server.router.HandleFunc("/endpoints", server.endpoints)
	server.router.HandleFunc("/system", server.systemStatus)
	server.router.HandleFunc("/api/v1/pubkey", server.publicKey)
	server.router.HandleFunc("/api/v1/register", registrationService.Register)
	server.router.HandleFunc("/api/v1/login", jsonWebTokenService.GenerateToken)
	server.endpointList = append(server.endpointList, "/api/v1/register")
	server.endpointList = append(server.endpointList, "/api/v1/login")

	server.router.HandleFunc(fmt.Sprintf("%s/notification/{type}/{message}", baseFarmURI), server.sendNotification)
	server.router.HandleFunc(fmt.Sprintf("%s/notification/{type}/{message}/{priority}", baseFarmURI), server.sendNotification)

	farmStateURI := fmt.Sprintf("%s/state", baseURI)
	server.router.HandleFunc(farmStateURI, server.state)
	server.endpointList = append(server.endpointList, farmStateURI)

	// farmConfigURI := fmt.Sprintf("%s/config", baseURI)
	// server.router.HandleFunc(farmConfigURI, server.config)
	// server.endpointList = append(server.endpointList, farmConfigURI)

	//server.router.HandleFunc("/maint/{mode}", server.MaintenanceMode).Methods("GET")
	farmMaintModeURI := fmt.Sprintf("%s/maint/{mode}", baseFarmURI)
	server.router.HandleFunc(farmMaintModeURI, server.MaintenanceMode).Methods("GET")
	server.endpointList = append(server.endpointList, farmMaintModeURI)

	// REST Handlers - JWT authentication required
	for _, restService := range server.restServices {
		endpoints := restService.RegisterEndpoints(server.router, baseURI, baseFarmURI)
		for _, endpoint := range endpoints {
			server.app.Logger.Debugf("Loading REST resource: %s", endpoint)
			server.endpointList = append(server.endpointList, endpoint)
		}
	}

	endpoint := fmt.Sprintf("%s/events", baseFarmURI)
	server.router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(server.events)),
	))
	server.endpointList = append(server.endpointList, endpoint)

	endpoint = fmt.Sprintf("%s/events/{page}", baseFarmURI)
	server.router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(server.eventsPage)),
	))
	server.endpointList = append(server.endpointList, endpoint)

	// /virtual
	/*
		server.router.Handle("/api/v1/virtual/{vdevice}/{metric}/{value}", negroni.New(
			negroni.HandlerFunc(jsonWebTokenService.Validate),
			negroni.Wrap(http.HandlerFunc(server.setVirtualMetric)),
		)).Methods("GET")
	*/

	// Websocket hubs
	notificationHub := websocket.NewNotificationHub(server.app.Logger, notificationService)
	go notificationHub.Run()

	// Websocket handlers
	endpoint = fmt.Sprintf("%s/notifications", baseFarmURI)
	server.router.Handle(endpoint, negroni.New(
		negroni.HandlerFunc(jsonWebTokenService.Validate),
		negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler := websocket.NewNotificationHandler(server.app.Logger, notificationHub, jsonWebTokenService)
			handler.OnConnect(w, r)
		})),
	))
	//server.endpointList = append(server.endpointList, endpoint)

	for _, farmService := range server.registry.GetFarmServices() {
		farmHub := websocket.NewFarmHub(server.app.Logger, notificationService, farmService)
		go farmHub.Run()
		endpoint = fmt.Sprintf("%s/farmticker/{farmID}", baseURI)
		server.router.Handle(endpoint, negroni.New(
			negroni.HandlerFunc(jsonWebTokenService.Validate),
			negroni.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler := websocket.NewFarmHandler(server.app.Logger, farmHub, notificationService, jsonWebTokenService)
				handler.OnConnect(w, r)
			})),
		))
		//server.endpointList = append(server.endpointList, endpoint)
	}

	sort.Strings(server.endpointList)

	server.app.Logger.Debugf("Loaded %d REST services", len(server.endpointList))

	// Register static content endpoint after REST and websocket to avoid being clobbered
	server.router.PathPrefix("/").Handler(fs)
	http.Handle("/", server.router)

	if server.app.SSLFlag {

		server.app.Logger.Debugf("Starting web services on TLS port %d", server.app.WebPort)
		eventLogService.Create(server.eventType, fmt.Sprintf("Starting web server on TLS port %d", server.app.WebPort))

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

		err = server.httpServer.Serve(listener)
		if err != nil {
			server.app.Logger.Fatalf("[WebServer] Unable to start web server: %s", err.Error())
		}

	} else {

		server.app.Logger.Infof("Starting web services on port %d", server.app.WebPort)
		eventLogService.Create(server.eventType, fmt.Sprintf("Starting web services on port %d", server.app.WebPort))

		ipv4Listener, err := net.Listen("tcp4", sPort)
		if err != nil {
			log.Fatal(err)
		}

		server.app.DropPrivileges()

		err = server.httpServer.Serve(ipv4Listener)
		if err != nil {
			server.app.Logger.Fatalf("[WebServer] Unable to start web server: %s", err.Error())
		}
	}
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

// func (server *Webserver) config(w http.ResponseWriter, r *http.Request) {
// 	rest.NewJsonWriter().Write(w, http.StatusOK, server.app.Config)
// }

func (server *Webserver) endpoints(w http.ResponseWriter, r *http.Request) {
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
		//DeviceIndexLength:   server.app.DeviceIndex.Len(),
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
		server.sendNotFound(w, r, fmt.Errorf("FarmID %d not found", farmID))
		return
	}

	farmState := farmService.GetState()
	if farmState != nil {
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

	server.app.Logger.Debugf("page %s requested", page)

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

func (server *Webserver) sendNotFound(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Not Found", http.StatusNotFound)
	fmt.Fprintln(w, err)
}

func (server *Webserver) sendBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Bad Request", http.StatusBadRequest)
	fmt.Fprintln(w, err)
}

func (server *Webserver) sendInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	fmt.Fprintln(w, err)
}
