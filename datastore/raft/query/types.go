package query

type PagerProcFunc[E any] func(entities []E) error

const (
	QUERY_TYPE_COUNT = iota
)

const (
	SORT_ASCENDING = iota
	SORT_DESCENDING
)

type PageQuery struct {
	Page      int
	PageSize  int
	SortOrder int
}

func NewPageQuery() PageQuery {
	return PageQuery{
		Page:     1,
		PageSize: 25}
}
