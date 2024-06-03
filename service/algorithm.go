package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
)

type AlgorithmServicer interface {
	Page(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.AlgorithmStruct], error)
}

type AlgorithmService struct {
	dao dao.AlgorithmDAO
	AlgorithmServicer
}

func NewAlgorithmService(dao dao.AlgorithmDAO) AlgorithmServicer {
	return &AlgorithmService{dao: dao}
}

func (service *AlgorithmService) Page(pageQuery query.PageQuery,
	CONSISTENCY_LEVEL int) (dao.PageResult[*config.AlgorithmStruct], error) {

	return service.dao.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}
