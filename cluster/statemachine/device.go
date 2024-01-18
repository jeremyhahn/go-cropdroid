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
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type DeviceConfigOnDiskStateMachine interface {
	CreateDeviceConfigOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type DeviceDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	DeviceConfigOnDiskStateMachine
}

func NewDeviceConfigOnDiskStateMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) DeviceConfigOnDiskStateMachine {

	return &DeviceDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *DeviceDiskKV) CreateDeviceConfigOnDiskStateMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *DeviceDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *DeviceDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *DeviceDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *DeviceDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *DeviceDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *DeviceDiskKV) Close() error {
	return d.diskKV.Close()
}

// Lookup expects the uint64 organization ID as the key
func (d *DeviceDiskKV) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var deviceConfig config.Device
		err = json.Unmarshal(v.([]byte), &deviceConfig)
		if err != nil {
			d.logger.Errorf("[DeviceDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &deviceConfig, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[DeviceDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		if query == QUERY_TYPE_WILDCARD {
			db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
			iter := db.db.NewIter(db.ro)
			defer iter.Close()
			values := make([]KVData, 0)
			for iter.First(); iter.Valid(); iter.Next() {
				kv := KVData{
					Key: iter.Key(),
					Val: iter.Value(),
				}
				values = append(values, kv)
			}
			devices := make([]*config.Device, 0)
			for _, deviceKV := range values {
				if string(deviceKV.Key) == appliedIndexKey {
					continue
				}
				var device config.Device
				if err := json.Unmarshal(deviceKV.Val, &device); err != nil {
					return nil, err
				}
				devices = append(devices, &device)
			}
			return devices, nil
		}
		return nil, nil
	}
	return nil, datastore.ErrNotFound
}

// Updates and deletes items from the ondisk state machine
func (d *DeviceDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[DeviceConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var deviceConfig config.Device
		err = json.Unmarshal(proposal.Data, &deviceConfig)
		if err != nil {
			d.logger.Errorf("[DeviceConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			Key: []byte(fmt.Sprint(deviceConfig.GetID())),
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
				d.logger.Errorf("[DeviceConfigMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		//d.deviceConfigChangeChan <- &deviceConfig
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
