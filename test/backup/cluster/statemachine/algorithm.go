//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"io"
	"sync/atomic"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type AlgorithmConfigMachine interface {
	CreateAlgorithmConfigMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type AlgorithmDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	AlgorithmConfigMachine
}

func NewAlgorithmConfigMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) AlgorithmConfigMachine {

	return &AlgorithmDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *AlgorithmDiskKV) CreateAlgorithmConfigMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *AlgorithmDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *AlgorithmDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *AlgorithmDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *AlgorithmDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *AlgorithmDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *AlgorithmDiskKV) Close() error {
	return d.diskKV.Close()
}

func (d *AlgorithmDiskKV) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var algorithmConfig config.Algorithm
		err = json.Unmarshal(v.([]byte), &algorithmConfig)
		if err != nil {
			d.logger.Errorf("[AlgorithmDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &algorithmConfig, err

	case []uint8: // json serialized struct
		var pageQuery query.PageQuery
		err := json.Unmarshal([]byte(string(key.([]uint8))), &pageQuery)
		if err != nil {
			d.logger.Error(err)
			return nil, err
		}
		return d.GetPage(pageQuery) // else config.KeyValueEntity

	}

	return nil, datastore.ErrNotFound
}

func (d *AlgorithmDiskKV) GetPage(pageQuery query.PageQuery) (interface{}, error) {

	pageResult := dao.PageResult[*config.Algorithm]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}

	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageQuery.PageSize

	db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
	iter := db.db.NewIter(db.ro)
	defer iter.Close()

	i := 0
	var entities = make([]*config.Algorithm, 0)
	if pageQuery.SortOrder == query.SORT_ASCENDING {
		for iter.First(); iter.Valid(); iter.Next() {
			if i < offset {
				i++
				continue
			}
			if i == offset+pageQuery.PageSize {
				pageResult.HasMore = iter.Next() // peek the next record
				break
			}
			i++
			key := iter.Key()
			value := iter.Value()
			if string(key) == "applied_index" {
				// This is the dragonboat applied_index field
				continue
			}
			var entity *config.Algorithm
			err := json.Unmarshal(value, &entity)
			if err != nil {
				d.logger.Errorf("[Page] Error: %s\n", err)
				return nil, err
			}
			entities = append(entities, entity)
		}
	} else {
		for iter.Last(); iter.Valid(); iter.Prev() {
			if i < offset {
				i++
				continue
			}
			if i == offset+pageQuery.PageSize {
				pageResult.HasMore = iter.Next()
				break
			}
			i++
			key := iter.Key()
			value := iter.Value()
			if string(key) == "applied_index" {
				continue
			}
			var entity *config.Algorithm
			err := json.Unmarshal(value, &entity)
			if err != nil {
				d.logger.Errorf("[Page] Error: %s\n", err)
				return nil, err
			}
			entities = append(entities, entity)
		}
	}
	pageResult.Entities = entities
	return pageResult, nil
}

func (d *AlgorithmDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[AlgorithmConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var algorithmConfig config.Algorithm
		err = json.Unmarshal(proposal.Data, &algorithmConfig)
		if err != nil {
			d.logger.Errorf("[AlgorithmConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			Key: d.idGenerator.Uint64Bytes(algorithmConfig.ID),
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
				d.logger.Errorf("[AlgorithmConfigMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		//d.algorithmConfigChangeChan <- &algorithmConfig
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
