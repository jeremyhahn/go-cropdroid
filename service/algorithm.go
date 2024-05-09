package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
)

type AlgorithmService interface {
	GetPage(CONSISTENCY_LEVEL, page, pageSize int) ([]config.Algorithm, error)
}

type DefaultAlgorithmService struct {
	dao dao.AlgorithmDAO
	AlgorithmService
}

func NewAlgorithmService(dao dao.AlgorithmDAO) AlgorithmService {
	return &DefaultAlgorithmService{dao: dao}
}

func (service *DefaultAlgorithmService) GetPage(CONSISTENCY_LEVEL, page, pageSize int) ([]config.Algorithm, error) {
	entities, err := service.dao.GetPage(common.CONSISTENCY_LOCAL, page, pageSize)
	if err != nil {
		return nil, err
	}
	return entities, nil
}
