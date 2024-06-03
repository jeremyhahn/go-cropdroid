package router

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/response"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
)

type RegistrationRouter struct {
	jsonWebTokenService     rest.JsonWebTokenServicer
	registrationRestService rest.RegistrationRestServicer
	WebServiceRouter
}

// Creates a new web service registration router
func NewRegistrationRouter(
	app *app.App,
	userService service.UserServicer,
	jsonWriter response.HttpWriter) WebServiceRouter {

	return &RegistrationRouter{
		registrationRestService: rest.NewRegistrationRestService(
			app,
			userService,
			jsonWriter)}
}

// Registers all of the registration endpoints at the root of the webservice (/api/v1)
func (registrationRouter *RegistrationRouter) RegisterRoutes(router *mux.Router, baseURI string) []string {
	return []string{
		registrationRouter.register(router, baseURI),
		registrationRouter.activate(router, baseURI)}
}

// @Summary New user registration
// @Description Creates a new user registration and sends an activation email
// @Tags Registration
// @Produce  json
// @Param UserCredentials body service.UserCredentials true "UserCredentials struct"
// @Success 200 {object} rest.RegistrationResponse
// @Failure 500 {object} response.WebServiceResponse
// @Router /register [post]
func (registrationRouter *RegistrationRouter) register(router *mux.Router, baseURI string) string {
	register := fmt.Sprintf("%s/register", baseURI)
	router.HandleFunc(register, registrationRouter.registrationRestService.Register)
	return register
}

// @Summary Activte pending registration
// @Description Creates a new user and deletes the pending registration
// @Tags Registration
// @Produce  json
// @Param 	token	path	string	true	"string valid"	minlength(1)	maxlength(20)
// @Success 200 {object} rest.RegistrationResponse
// @Failure 500 {object} response.WebServiceResponse
// @Router /register/activate/{token} [get]
func (registrationRouter *RegistrationRouter) activate(router *mux.Router, baseURI string) string {
	activate := fmt.Sprintf("%s/register/activate/{token}", baseURI)
	router.HandleFunc(activate, registrationRouter.registrationRestService.Activate)
	return activate
}
