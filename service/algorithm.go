package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
)

type AlgorithmService interface {
	GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Algorithm], error)
}

type DefaultAlgorithmService struct {
	dao dao.AlgorithmDAO
	AlgorithmService
}

func NewAlgorithmService(dao dao.AlgorithmDAO) AlgorithmService {
	return &DefaultAlgorithmService{dao: dao}
}

func (service *DefaultAlgorithmService) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.Algorithm], error) {
	return service.dao.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}
