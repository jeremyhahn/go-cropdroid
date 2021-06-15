package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	mock.Mock
	http.Client
}

func (m *MockHttpClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}
