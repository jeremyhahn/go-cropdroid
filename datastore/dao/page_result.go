package dao

import "github.com/jeremyhahn/go-cropdroid/datastore/raft/query"

type PageResult[E any] struct {
	Entities []E  `yaml:"entities" json:"entities"`
	Page     int  `yaml:"page" json:"page"`
	PageSize int  `yaml:"pageSize" json:"pageSize"`
	HasMore  bool `yaml:"has_more" json:"has_more"`
}

func NewPageResult[E any]() PageResult[E] {
	return PageResult[E]{Entities: make([]E, 0)}
}

func NewPageResultFromQuery[E any](q query.PageQuery) PageResult[E] {
	return PageResult[E]{
		Entities: make([]E, q.PageSize),
		Page:     q.Page,
		PageSize: q.PageSize}
}
