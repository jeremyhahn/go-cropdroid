package middleware

import (
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/service"
)

type JsonWebTokenMiddleware interface {
	Validate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
	CreateSession(w http.ResponseWriter, r *http.Request) (service.Session, error)
}

type AuthMiddleware interface {
	GenerateToken(w http.ResponseWriter, req *http.Request)
	RefreshToken(w http.ResponseWriter, req *http.Request)
}
