package service

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
)

type GenericServicer[E any] interface {
	Save(entity E) error
	GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[E], error)
}

type GenericService[E any] struct {
	dao dao.GenericDAO[E]
	GenericServicer[E]
}

func NewGenericService[E any](dao dao.GenericDAO[E]) GenericServicer[E] {
	return &GenericService[E]{dao: dao}
}

func (service *GenericService[E]) Save(entity E) error {
	return service.dao.Save(entity)
}

func (service *GenericService[E]) Get(id uint64, CONSISTENCY_LEVEL int) (E, error) {
	return service.dao.Get(id, CONSISTENCY_LEVEL)
}

func (service *GenericService[E]) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[E], CONSISTENCY_LEVEL int) error {

	return service.dao.ForEachPage(pageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
}

func (service *GenericService[E]) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[E], error) {
	return service.dao.GetPage(pageQuery, common.CONSISTENCY_LOCAL)
}

func (service *GenericService[E]) Delete(entity E) error {
	return service.dao.Delete(entity)
}

func (service *GenericService[E]) Count(CONSISTENCY_LEVEL int) (int64, error) {
	return service.dao.Count(CONSISTENCY_LEVEL)
}
