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

type DeviceSettingMachine interface {
	CreateDeviceSettingMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type DeviceSettingDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	DeviceSettingMachine
}

func NewDeviceSettingConfigMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) DeviceSettingMachine {

	return &DeviceSettingDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *DeviceSettingDiskKV) CreateDeviceSettingMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *DeviceSettingDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *DeviceSettingDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *DeviceSettingDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *DeviceSettingDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *DeviceSettingDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *DeviceSettingDiskKV) Close() error {
	return d.diskKV.Close()
}

// Lookup expects the uint64 organization ID as the key
func (d *DeviceSettingDiskKV) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var deviceSetting config.DeviceSetting
		err = json.Unmarshal(v.([]byte), &deviceSetting)
		if err != nil {
			d.logger.Errorf("[DeviceSettingDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &deviceSetting, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[DeviceSettingDiskKV.Lookup] Error: %s\n", err)
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
			deviceSettings := make([]config.DeviceSetting, 0)
			for _, deviceKV := range values {
				if string(deviceKV.Key) == appliedIndexKey {
					continue
				}
				var setting config.DeviceSetting
				if err := json.Unmarshal(deviceKV.Val, &setting); err != nil {
					return nil, err
				}
				deviceSettings = append(deviceSettings, setting)
			}
			return deviceSettings, nil
		}
		return nil, nil
	}
	return nil, datastore.ErrNotFound
}

// Updates and deletes items from the ondisk state machine
func (d *DeviceSettingDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[DeviceSettingMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var deviceSetting config.DeviceSetting
		err = json.Unmarshal(proposal.Data, &deviceSetting)
		if err != nil {
			d.logger.Errorf("[DeviceSettingMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			Key: []byte(fmt.Sprint(deviceSetting.GetID())),
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
				d.logger.Errorf("[DeviceSettingMachine.Update] Error: %s\n", err)
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
