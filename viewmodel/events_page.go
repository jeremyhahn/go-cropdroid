package viewmodel

import "github.com/jeremyhahn/cropdroid/datastore/gorm/entity"

type EventsPage struct {
	Events []entity.EventLog `json:"events"`
	Page   int64             `json:"page"`
	Size   int64             `json:"size"`
	Count  int64             `json:"count"`
	Start  int64             `json:"start"`
	End    int64             `json:"end"`
}
