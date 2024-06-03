package router

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/middleware"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type GoogleRouter struct {
	googleAuthService rest.GoogleRestServicer
	WebServiceRouter
}

// Creates a new web service google router
func NewGoogleRouter(
	authService service.AuthServicer,
	middleware middleware.AuthMiddleware,
	httpWriter response.HttpWriter) WebServiceRouter {

	return &GoogleRouter{
		googleAuthService: rest.NewGoogleRestService(
			authService,
			middleware,
			httpWriter)}
}

// Registers all of the google endpoints at the root of the webservice (/api/v1)
func (googleRouter *GoogleRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		googleRouter.login(router, baseURI)}
}

// @Summary Google Authentication
// @Description Authenticates the user using the Google Auth API
// @Tags Authentication
// @Produce  json
// @Param UserCredentials body service.UserCredentials true "UserCredentials struct"
// @Success 200
// @Failure 400 {object} response.WebServiceResponse
// @Router /google/login [post]
// @Security JWT
func (googleRouter *GoogleRouter) login(router *mux.Router, baseFarmURI string) string {
	endpoint := fmt.Sprintf("%s/google/login", baseFarmURI)
	router.HandleFunc(endpoint, googleRouter.googleAuthService.Login)
	return endpoint
}
