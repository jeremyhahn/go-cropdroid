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
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type FarmConfigMachine interface {
	CreateFarmConfigMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type FarmDiskKV struct {
	logger               *logging.Logger
	idGenerator          util.IdGenerator
	farmConfigChangeChan chan config.Farm
	diskKV               DiskKV
	FarmConfigMachine
}

func NewFarmConfigMachine(logger *logging.Logger, idGenerator util.IdGenerator,
	clusterID uint64, dbPath string, farmConfigChangeChan chan config.Farm) FarmConfigMachine {

	return &FarmDiskKV{
		logger:               logger,
		idGenerator:          idGenerator,
		farmConfigChangeChan: farmConfigChangeChan,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID}}
}

func (d *FarmDiskKV) CreateFarmConfigMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine {
	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *FarmDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *FarmDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *FarmDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *FarmDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *FarmDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *FarmDiskKV) Close() error {
	return d.diskKV.Close()
}

func (d *FarmDiskKV) Lookup(key interface{}) (interface{}, error) {
	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var farmConfig config.Farm
		err = json.Unmarshal(v.([]byte), &farmConfig)
		if err != nil {
			d.logger.Errorf("[FarmDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &farmConfig, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[FarmDiskKV.Lookup] Error: %s\n", err)
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
			farms := make([]*config.Farm, 0)
			for _, farmKV := range values {
				if string(farmKV.Key) == appliedIndexKey {
					continue
				}
				var farm config.Farm
				if err := json.Unmarshal(farmKV.Val, &farm); err != nil {
					return nil, err
				}
				farms = append(farms, &farm)
			}
			return farms, nil
		}
		return nil, nil
	}
	return nil, ErrUnsupportedQuery
}

func (d *FarmDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[FarmConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var farmConfig config.Farm
		err = json.Unmarshal(proposal.Data, &farmConfig)
		if err != nil {
			d.logger.Errorf("[FarmConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			//Key: []byte(fmt.Sprint(farmConfig.GetID())),
			Key: d.idGenerator.Uint64Bytes(farmConfig.GetID()),
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
				d.logger.Errorf("[FarmConfigMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		//d.farmConfigChangeChan <- &farmConfig
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
