package restservice

import (
	"net/http"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/webservice/v1/rest"
	"github.com/stretchr/testify/mock"
)

type MockJsonWebTokenService struct {
	app *app.App
	mock.Mock
	rest.JsonWebTokenServicer
}

func NewMockJsonWebTokenService() rest.JsonWebTokenServicer {
	return &MockJsonWebTokenService{}
}

func (m *MockJsonWebTokenService) SetApp(app *app.App) {
	m.app = app
}

func (m *MockJsonWebTokenService) CreateSession(w http.ResponseWriter, r *http.Request) (service.Session, error) {
	session := service.CreateSession(
		m.app.Logger,
		[]service.OrganizationClaim{},
		[]service.FarmClaim{},
		nil,
		0,
		0,
		common.CONSISTENCY_LOCAL,
		model.NewUser())
	args := m.Called(w, r)
	return session, args.Error(1)
}
