package dao

type PageResult[E any] struct {
	Entities []E  `yaml:"entities" json:"entities"`
	Page     int  `yaml:"page" json:"page"`
	PageSize int  `yaml:"pageSize" json:"pageSize"`
	HasMore  bool `yaml:"has_more" json:"has_more"`
}

func NewPageResult[E any]() PageResult[E] {
	return PageResult[E]{Entities: make([]E, 0)}
}
