package rest

import "github.com/gorilla/mux"

type RestService interface {
	RegisterEndpoints(router *mux.Router, baseURI, baseFarmURI string) []string
}

type RestServiceRegistry interface {
	GetRestServices() []RestService
}
