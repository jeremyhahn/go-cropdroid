package service

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
)

type AlgorithmService interface {
	GetAll() ([]config.Algorithm, error)
}

type DefaultAlgorithmService struct {
	dao dao.AlgorithmDAO
	AlgorithmService
}

func NewAlgorithmService(dao dao.AlgorithmDAO) AlgorithmService {
	return &DefaultAlgorithmService{dao: dao}
}

func (service *DefaultAlgorithmService) GetAll() ([]config.Algorithm, error) {
	entities, err := service.dao.GetAll()
	if err != nil {
		return nil, err
	}
	return entities, nil
}
