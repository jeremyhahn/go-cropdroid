package restservice

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockAlgorithmRestService struct {
	mock.Mock
}

func NewMockAlgorithmRestService() *MockAlgorithmRestService {
	return &MockAlgorithmRestService{}
}

func (dao *MockAlgorithmRestService) Page(w http.ResponseWriter, r *http.Request) {
	dao.Called(w, r)
}
