package store

import (
	"sync"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/config/store"
)

type GormFarmConfigStore struct {
	farmDAO      dao.FarmDAO
	cachedConfig map[uint64]config.Farm
	mutex        *sync.RWMutex
	store.FarmConfigStorer
}

func NewGormFarmConfigStore(farmDAO dao.FarmDAO, len int) store.FarmConfigStorer {
	return &GormFarmConfigStore{
		farmDAO:      farmDAO,
		mutex:        &sync.RWMutex{},
		cachedConfig: make(map[uint64]config.Farm, 0)}
}

func (s *GormFarmConfigStore) Len() int {
	count, _ := s.farmDAO.Count()
	return int(count)
}

func (s *GormFarmConfigStore) Cache(farmID uint64, c config.FarmConfig) {
	if c != nil {
		s.mutex.Lock()
		s.cachedConfig[farmID] = *c.(*config.Farm)
		s.mutex.Unlock()
	}
}

func (s *GormFarmConfigStore) Put(farmID uint64, c config.FarmConfig) error {
	farmConfig := c.(*config.Farm)
	// s.Cache(farmID, farmConfig)
	return s.farmDAO.Save(farmConfig)
}

func (s *GormFarmConfigStore) Get(farmID uint64, CONSISTENCY_LEVEL int) (config.FarmConfig, error) {
	// if CONSISTENCY_LEVEL == common.CONSISTENCY_CACHED {
	// 	if config, ok := s.cachedConfig[farmID]; ok {
	// 		return &config, nil
	// 	}
	// }
	config, err := s.farmDAO.Get(farmID, CONSISTENCY_LEVEL)
	// s.Cache(farmID, config)
	return config, err
}

func (s *GormFarmConfigStore) GetByIds(farmIds []uint64, CONSISTENCY_LEVEL int) []config.FarmConfig {
	configs, _ := s.farmDAO.GetByIds(farmIds, CONSISTENCY_LEVEL)
	return configs
}

func (s *GormFarmConfigStore) GetAll() []config.FarmConfig {
	configs, _ := s.farmDAO.GetAll()
	return configs
}
