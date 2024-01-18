//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type EventLogOnDiskStateMachine interface {
	CreateEventLogOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type EventLogDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	diskKV      DiskKV
	EventLogOnDiskStateMachine
}

func NewEventLogOnDiskStateMachine(logger *logging.Logger, idGenerator util.IdGenerator,
	clusterID uint64, dbPath string) EventLogOnDiskStateMachine {

	eventLogClusterID := idGenerator.NewID(fmt.Sprintf("%d-%s", clusterID, "eventlog"))

	return &EventLogDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: eventLogClusterID}}
}

func (d *EventLogDiskKV) CreateEventLogOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine {
	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	//d.diskKV.clusterID = clusterID
	d.diskKV.clusterID = d.idGenerator.NewID(fmt.Sprintf("%d-%s", clusterID, "eventlog"))
	d.diskKV.nodeID = nodeID
	return d
}

func (d *EventLogDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *EventLogDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *EventLogDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *EventLogDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *EventLogDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *EventLogDiskKV) Close() error {
	return d.diskKV.Close()
}

func (d *EventLogDiskKV) Lookup(key interface{}) (interface{}, error) {
	switch t := key.(type) {

	case int:
		query := key.(int)
		if query == QUERY_TYPE_COUNT {
			d.logger.Debug("[EventLogDiskKV.Lookup] QUERY_TYPE_COUNT")
			db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
			iter := db.db.NewIter(db.ro)
			defer iter.Close()
			i := 0
			for iter.First(); iter.Valid(); iter.Next() {
				i++
			}
			return i, nil
		} else if query == QUERY_TYPE_WILDCARD {
			db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
			iter := db.db.NewIter(db.ro)
			defer iter.Close()
			values := make([]KVData, 0)
			i := 0
			for iter.First(); iter.Valid(); iter.Next() {
				key := make([]byte, len(iter.Key()))
				copy(key, iter.Key())

				value := make([]byte, len(iter.Value()))
				copy(value, iter.Value())

				values = append(values, KVData{
					Key: key,
					Val: value})

				i++
			}
			records := make([]entity.EventLog, i-1)
			for _, eventLogDataKV := range values {
				if string(eventLogDataKV.Key) == appliedIndexKey {
					continue
				}
				var eventLogRecord entity.EventLog
				if err := json.Unmarshal(eventLogDataKV.Val, &eventLogRecord); err != nil {
					return nil, err
				}
				records = append(records, eventLogRecord)
			}
			return records, nil
		}

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var deviceData entity.EventLogEntity
		err = json.Unmarshal(v.([]byte), &deviceData)
		if err != nil {
			d.logger.Errorf("[EventLogDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &deviceData, err

	}
	return nil, ErrUnsupportedQuery
}

func (d *EventLogDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[EventLogMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var eventLog entity.EventLog
		err = json.Unmarshal(proposal.Data, &eventLog)
		if err != nil {
			d.logger.Errorf("[EventLogMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			//Key: []byte(fmt.Sprint(deviceData.GetID())),
			Key: d.idGenerator.Uint64Bytes(uint64(idx)),
			Val: proposal.Data}

		jsonDataKV, err := json.Marshal(kvdata)
		if err != nil {
			return nil, err
		}

		entry := sm.Entry{
			Index: e.Index,
			Cmd:   jsonDataKV}

		if proposal.Query == QUERY_TYPE_DELETE {
			err = d.diskKV.Delete(entry)
			if err != nil {
				d.logger.Errorf("[EventLogMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
