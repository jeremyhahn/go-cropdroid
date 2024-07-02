// @title CropDroid REST API
// @version v0.0.3
// @description This is the RESTful web servce for CropDroid.
// @termsOfService https://www.cropdroid.com/terms/

// @contact.name API Support
// @contact.url https://www.cropdroid.com/support
// @contact.email support@cropdroid.com

// @license.name GNU AFFERO GENERAL PUBLIC LICENSE
// @license.url https://www.gnu.org/licenses/agpl-3.0.txt

// @license.name Commercial
// @license.url https://www.cropdroid.com/licenses/commercial.txt

// @host localhost:8443
// @BasePath /api/v1
// schemes: [http, https, ws, wss]
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization

package webservice

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/websocket"

	v1 "github.com/jeremyhahn/go-cropdroid/webservice/v1"
)

var (
	ErrLoadTlsCerts = errors.New("unable to load TLS certificates")
	ErrBindPort     = errors.New("unable to bind to web service port")
)

type WebServerV1 struct {
	app                       *app.App
	config                    config.WebService
	baseURI                   string
	eventType                 string
	endpointList              []string
	routerMutex               *sync.Mutex
	router                    *mux.Router
	httpServer                *http.Server
	mapperRegistry            mapper.MapperRegistry
	serviceRegistry           service.ServiceRegistry
	restServiceRegistry       rest.RestServiceRegistry
	middleware                middleware.JsonWebTokenMiddleware
	systemEventLogService     service.EventLogServicer
	notificationService       service.NotificationServicer
	farmWebSocketHandler      rest.FarmWebSocketRestServicer
	farmHubMutex              *sync.RWMutex
	farmHubs                  map[uint64]*websocket.FarmHub
	notificationHubs          map[uint64]*websocket.NotificationHub
	notificationHubMutex      *sync.RWMutex
	farmTickerProvisionerChan chan uint64
	closeChan                 chan bool
}

func NewWebServerV1(
	app *app.App,
	mapperRegistry mapper.MapperRegistry,
	serviceRegistry service.ServiceRegistry,
	restServiceRegistry rest.RestServiceRegistry,
	farmTickerProvisionerChan chan uint64) *WebServerV1 {

	webserver := &WebServerV1{
		app:                       app,
		baseURI:                   "/api/v1",
		eventType:                 "WebServer",
		endpointList:              make([]string, 0),
		routerMutex:               &sync.Mutex{},
		router:                    mux.NewRouter().StrictSlash(true),
		mapperRegistry:            mapperRegistry,
		serviceRegistry:           serviceRegistry,
		restServiceRegistry:       restServiceRegistry,
		middleware:                restServiceRegistry.JsonWebTokenService(),
		systemEventLogService:     serviceRegistry.GetEventLogService(0),
		notificationService:       serviceRegistry.GetNotificationService(),
		farmHubMutex:              &sync.RWMutex{},
		farmHubs:                  make(map[uint64]*websocket.FarmHub),
		notificationHubs:          make(map[uint64]*websocket.NotificationHub),
		notificationHubMutex:      &sync.RWMutex{},
		farmTickerProvisionerChan: farmTickerProvisionerChan,
		closeChan:                 make(chan bool, 1)}

	webserver.httpServer = &http.Server{
		ReadTimeout:  common.HTTP_SERVER_READ_TIMEOUT,
		WriteTimeout: common.HTTP_SERVER_WRITE_TIMEOUT,
		IdleTimeout:  common.HTTP_SERVER_IDLE_TIMEOUT,
		Handler:      webserver.router}

	// Manages all farm websocket connection pools / hubs
	webserver.farmWebSocketHandler = rest.NewFarmWebSocketRestService(
		app.Logger,
		webserver.farmHubs,
		webserver.farmHubMutex,
		webserver.notificationHubs,
		webserver.notificationHubMutex,
		webserver.serviceRegistry,
		webserver.middleware,
		response.NewResponseWriter(app.Logger, nil))

	return webserver
}

func (server *WebServerV1) Run() {

	server.buildRoutes()

	fs := http.FileServer(http.Dir(common.HTTP_PUBLIC_HTML))
	server.router.PathPrefix("/").Handler(fs)
	http.Handle("/", server.httpServer.Handler)

	if server.app.WebService.TLSPort > 0 {
		go server.startHttps()
	} else {
		go server.startHttp()
	}

	server.app.DropPrivileges()

	<-server.closeChan
}

func (server *WebServerV1) RunProvisionerConsumer() {
	for {
		farmID := <-server.farmTickerProvisionerChan
		server.app.Logger.Warningf("Web services rebuilding routes for farm: %d", farmID)
		server.buildRoutes()
	}
}

func (server WebServerV1) Shutdown() {
	server.app.Logger.Info("Web services shutting down")
	server.closeChan <- true
	close(server.closeChan)
	close(server.farmTickerProvisionerChan)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.httpServer.Shutdown(ctx)
	cancel()
}

func (server *WebServerV1) startHttp() {

	sWebPort := fmt.Sprintf(":%d", server.app.WebService.Port)

	insecureWebServicesMsg := fmt.Sprintf(
		"Starting insecure web services on plain-text HTTP port %s (not recommended)", sWebPort)
	server.app.Logger.Infof(insecureWebServicesMsg)
	server.systemEventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, server.eventType, insecureWebServicesMsg)

	ipv4Listener, err := net.Listen("tcp4", sWebPort)
	if err != nil {
		log.Fatal(err)
	}

	//err = http.Serve(ipv4Listener, server.router)
	err = server.httpServer.Serve(ipv4Listener)
	if err != nil {
		server.app.Logger.Fatalf("Unable to start web services: %s", err.Error())
	}
}

func (server *WebServerV1) startHttps() {

	sWebPort := fmt.Sprintf(":%d", server.app.WebService.Port)
	sTlsPort := fmt.Sprintf(":%d", server.app.WebService.TLSPort)

	message := fmt.Sprintf("Starting secure web services on TLS port %s", sTlsPort)
	server.app.Logger.Debugf(message)
	server.systemEventLogService.Create(0, common.CONTROLLER_TYPE_SERVER, server.eventType, message)

	server.app.Logger.Info("retrieving server private PEM key from cert store")
	privKeyPEM, err := server.app.CA.CertStore().PrivKeyPEM(server.app.Domain)
	if err != nil {
		server.app.Logger.Fatal(err)
	}

	server.app.Logger.Info("retrieving server public PEM key from cert store")
	certPEM, err := server.app.CA.PEM(server.app.Domain)
	if err != nil {
		server.app.Logger.Fatal(err)
	}

	server.app.Logger.Info("creating server x509 key pair")
	serverCertificate, err := tls.X509KeyPair(certPEM, privKeyPEM)
	if err != nil {
		server.app.Logger.Fatal(err)
	}

	rootCertPool, err := server.app.CA.CertStore().TrustedRootCertPool(server.app.CAConfig.AutoImportIssuingCA)
	if err != nil {
		server.app.Logger.Fatal(err)
	}

	tlsconf := &tls.Config{
		// ClientCAs:    clientCertPool,
		// ClientAuth:   tls.RequireAndVerifyClientCert,
		//GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		// Always get latest certificate
		// cert, err := tls.X509KeyPair(newPubPEM, newPrivPEM)
		// if err != nil {
		// 	return nil, err
		// }
		// return &cert, nil
		//},
		RootCAs:      rootCertPool,
		Certificates: []tls.Certificate{serverCertificate},
	}

	// Set up an HTTPS client configured to trust the CA
	//
	// caPEM, err := server.app.CA.PEM("ca")
	// if err != nil {
	// 	server.app.Logger.Fatalf(ErrLoadTlsCerts.Error())
	// }
	//
	// certpool := x509.NewCertPool()
	// certpool.AppendCertsFromPEM(caPEM)
	// clientTLSConf ;= &tls.Config{
	// 	RootCAs: certpool,
	// }

	// transport := &http.Transport{
	// 	TLSClientConfig: clientTLSConf,
	// }
	// http := http.Client{
	// 	Transport: transport,
	// }

	server.httpServer.TLSConfig = tlsconf

	tlsListener, err := tls.Listen("tcp4", sTlsPort, tlsconf)
	if err != nil {
		server.app.Logger.Fatalf("%s: %d", ErrBindPort, server.app.WebService.TLSPort)
	}

	if server.app.RedirectHttpToHttps {
		server.app.Logger.Debugf("Redirecting HTTP traffic to HTTPS")
		go http.ListenAndServe(sWebPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+sWebPort+r.URL.String(), http.StatusMovedPermanently)
		}))
	} else {
		go server.startHttp()
	}

	err = server.httpServer.Serve(tlsListener)
	if err != nil {
		server.app.Logger.Fatalf("Unable to start TLS web server: %s", err.Error())
	}
}

func (server *WebServerV1) buildRoutes() {
	muxRouter := mux.NewRouter().StrictSlash(true)
	responseWriter := response.NewResponseWriter(server.app.Logger, nil)
	endpointList := v1.NewRouterV1(
		server.app,
		server.mapperRegistry,
		server.serviceRegistry,
		server.restServiceRegistry,
		server.farmWebSocketHandler,
		muxRouter,
		responseWriter).RegisterRoutes(muxRouter, server.baseURI)
	server.routerMutex.Lock()
	server.router = muxRouter
	copy(server.endpointList, endpointList)
	server.httpServer.Handler = server.router
	server.routerMutex.Unlock()
}

// func (server *WebServer) walkRoutes() {
// 	err := server.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
// 		pathTemplate, err := route.GetPathTemplate()
// 		if err == nil {
// 			server.app.Logger.Debug("ROUTE:", pathTemplate)
// 		}
// 		pathRegexp, err := route.GetPathRegexp()
// 		if err == nil {
// 			server.app.Logger.Debug("Path regexp:", pathRegexp)
// 		}
// 		queriesTemplates, err := route.GetQueriesTemplates()
// 		if err == nil {
// 			server.app.Logger.Debug("Queries templates:", strings.Join(queriesTemplates, ","))
// 		}
// 		queriesRegexps, err := route.GetQueriesRegexp()
// 		if err == nil {
// 			server.app.Logger.Debug("Queries regexps:", strings.Join(queriesRegexps, ","))
// 		}
// 		methods, err := route.GetMethods()
// 		if err == nil {
// 			server.app.Logger.Debug("Methods:", strings.Join(methods, ","))
// 		}
// 		server.app.Logger.Debug("")
// 		return nil
// 	})
// 	if err != nil {
// 		server.app.Logger.Errorf("[WebServer.WalkRoutes] Error: err", err)
// 	}
// }

// func (server *WebServer) MaintenanceMode(w http.ResponseWriter, r *http.Request) {

// 	//server.systemEventLogService.Create(server.eventType,
// 	//	fmt.Sprintf("/maint/%d requested by %s", mode, server.clientIP(r)))

// 	params := mux.Vars(r)

// 	farmID, err := strconv.ParseUint(params["farmID"], 10, 64)
// 	if err != nil {
// 		response.NewResponseWriter(server.app.Logger).Error400(w, r, err)
// 		return
// 	}

// 	/*
// 		mode, err := strconv.Atoi(params["mode"])
// 		if err != nil {
// 			server.app.Logger.Error(err.Error())
// 			server.sendBadRequest(w, r, err)
// 			return
// 		}*/

// 	farmService := server.serviceRegistry.GetFarmService(farmID)
// 	if farmService == nil {
// 		server.app.Logger.Error(response.ErrFarmNotFound)
// 		return
// 	}
// 	farmState := farmService.GetState()
// 	if farmState == nil {
// 		response.NewResponseWriter(server.app.Logger).Error400(w, r, err)
// 		return
// 	}

// 	/*
// 		if mode == 0 {
// 			farmState.SetMaintenanceMode(false)
// 			server.app.FarmStore.Put(farmID, farmState)
// 		} else {
// 			farmState.SetMaintenanceMode(true)
// 			server.app.FarmStore.Put(farmID, farmState)
// 		}*/

// 	w.Header().Set("Content-Type", "application/json")
// 	json, _ := json.MarshalIndent(farmState, "", " ")
// 	fmt.Fprintln(w, string(json))
// }
