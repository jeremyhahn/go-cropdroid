package datastore

import "encoding/json"

type DatabaseChangefeed struct {
	Table      string                      `json:"table"`
	Key        int64                       `json:"key"`
	Value      interface{}                 `json:"value"`
	Updated    string                      `json:"updated"`
	Bytes      []byte                      `json:"-"`
	RawMessage map[string]*json.RawMessage `json:"-"`
	Changefeed `yaml:"-" json:"-"`
}

func (changefeed *DatabaseChangefeed) GetTable() string {
	return changefeed.Table
}

func (changefeed *DatabaseChangefeed) GetKey() int64 {
	return changefeed.Key
}

func (changefeed *DatabaseChangefeed) GetBytes() []byte {
	return changefeed.Bytes
}

func (changefeed *DatabaseChangefeed) GetRawMessage() map[string]*json.RawMessage {
	return changefeed.RawMessage
}

func (changefeed *DatabaseChangefeed) GetValue() interface{} {
	return changefeed.Value
}

func (changefeed *DatabaseChangefeed) GetUpdated() string {
	return changefeed.Updated
}
