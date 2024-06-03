package service

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/stretchr/testify/mock"
)

type MockAlgorithmService struct {
	mock.Mock
	service.AlgorithmServicer
}

func NewMockAlgorithmService() service.AlgorithmServicer {
	return &MockAlgorithmService{}
}

func (service *MockAlgorithmService) Page(pageQuery query.PageQuery,
	CONSISTENCY_LEVEL int) (dao.PageResult[*config.AlgorithmStruct], error) {

	pageResult := dao.NewPageResultFromQuery[*config.AlgorithmStruct](pageQuery)
	for i := 0; i < pageQuery.PageSize; i++ {
		pageResult.Entities[i] = &config.AlgorithmStruct{
			ID:   uint64(i),
			Name: fmt.Sprintf("Test Algorithm %d", i)}
	}

	args := service.Called(pageQuery, CONSISTENCY_LEVEL)
	return pageResult, args.Error(1)
}
