//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type DeviceDataOnDiskStateMachine interface {
	CreateDeviceDataOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type DeviceDataDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	diskKV      DiskKV
	DeviceDataOnDiskStateMachine
}

func NewDeviceDataOnDiskStateMachine(logger *logging.Logger, idGenerator util.IdGenerator,
	clusterID uint64, dbPath string) DeviceDataOnDiskStateMachine {

	deviceDataID := idGenerator.NewID(fmt.Sprintf("%d-%s", clusterID, "devicedata"))

	return &DeviceDataDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: deviceDataID}}
}

func (d *DeviceDataDiskKV) CreateDeviceDataOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine {
	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	//d.diskKV.clusterID = clusterID
	d.diskKV.clusterID = d.idGenerator.NewID(fmt.Sprintf("%d-%s", clusterID, "devicedata"))
	d.diskKV.nodeID = nodeID
	return d
}

func (d *DeviceDataDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *DeviceDataDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *DeviceDataDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *DeviceDataDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *DeviceDataDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *DeviceDataDiskKV) Close() error {
	return d.diskKV.Close()
}

func (d *DeviceDataDiskKV) Lookup(key interface{}) (interface{}, error) {
	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var deviceData state.DeviceStateMap
		err = json.Unmarshal(v.([]byte), &deviceData)
		if err != nil {
			d.logger.Errorf("[DeviceDataDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &deviceData, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[DeviceDataDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		if query == QUERY_TYPE_WILDCARD {
			db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
			iter := db.db.NewIter(db.ro)
			defer iter.Close()
			values := make([]KVData, 0)
			for iter.First(); iter.Valid(); iter.Next() {

				key := make([]byte, len(iter.Key()))
				copy(key, iter.Key())

				value := make([]byte, len(iter.Value()))
				copy(value, iter.Value())

				values = append(values, KVData{
					Key: key,
					Val: value})
			}
			records := make([]state.DeviceStateMap, 0)
			for _, deviceDataKV := range values {
				if string(deviceDataKV.Key) == appliedIndexKey {
					continue
				}
				var deviceData state.DeviceStateMap
				if err := json.Unmarshal(deviceDataKV.Val, &deviceData); err != nil {
					return nil, err
				}
				records = append(records, deviceData)
			}
			return records, nil
		}
		return nil, nil
	}
	return nil, ErrUnsupportedQuery
}

func (d *DeviceDataDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[DeviceDataMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var deviceData state.DeviceStateMap
		err = json.Unmarshal(proposal.Data, &deviceData)
		if err != nil {
			d.logger.Errorf("[DeviceDataMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			//Key: []byte(fmt.Sprint(deviceData.GetID())),
			Key: d.idGenerator.Uint64Bytes(deviceData.GetID()),
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
				d.logger.Errorf("[DeviceDataMachine.Update] Error: %s\n", err)
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
