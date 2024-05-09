//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync/atomic"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/index"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type OnDiskStateMachine interface {
	CreateOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type GenericDiskKV[E interface{}] struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV[E]
	OnDiskStateMachine
}

func NewGenericOnDiskStateMachine[E interface{}](logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) OnDiskStateMachine {

	logger.Infof("Creatnig new %T Raft state machine", *new(E))

	return &GenericDiskKV[E]{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV[E]{
			dbPath:      dbPath,
			clusterID:   clusterID,
			nodeID:      nodeID,
			idGenerator: idGenerator}}
}

func (d *GenericDiskKV[E]) CreateOnDiskStateMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *GenericDiskKV[E]) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *GenericDiskKV[E]) Sync() error {
	return d.diskKV.Sync()
}

func (d *GenericDiskKV[E]) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *GenericDiskKV[E]) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *GenericDiskKV[E]) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *GenericDiskKV[E]) Close() error {
	return d.diskKV.Close()
}

func (d *GenericDiskKV[E]) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64: // get by id
		//id := d.idGenerator.Uint64Bytes(t)
		id := []byte(strconv.FormatUint(key.(uint64), 10))
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var entity E
		err = json.Unmarshal(v.([]byte), &entity)
		if err != nil {
			d.logger.Error(err)
			return nil, err
		}
		return entity, err

	case []uint8: // query
		var pageQuery query.Page
		err := json.Unmarshal([]byte(string(key.([]uint8))), &pageQuery)
		if err != nil {
			d.logger.Error(err)
			return nil, err
		}
		var entity E
		var i interface{} = entity
		_, ok := i.(config.TimeSeriesIndexeder)
		if ok {
			return d.GetPageUsingTimeSeriesIndex(pageQuery)
		}
		return d.GetPage(pageQuery) // else config.KeyValueEntity

	default:
		d.logger.Errorf("[keyType] %s\n", t)
	}
	return nil, datastore.ErrNotFound
}

func (d *GenericDiskKV[E]) GetPage(pageQuery query.Page) (interface{}, error) {

	var entities = make([]E, 0)

	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageQuery.PageSize

	db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
	iter := db.db.NewIter(db.ro)
	defer iter.Close()

	i := 0
	for iter.First(); iter.Valid(); iter.Next() {
		if i < offset {
			i++
			continue
		}
		if i == offset+pageQuery.PageSize {
			break
		}
		i++

		key := iter.Key()
		value := iter.Value()

		if string(key) == "applied_index" {
			// This is the dragonboat applied_index field
			continue
		}

		var entity E
		err := json.Unmarshal(value, &entity)
		if err != nil {
			d.logger.Errorf("[Page] Error: %s\n", err)
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

func (d *GenericDiskKV[E]) Update(ents []sm.Entry) ([]sm.Entry, error) {

	var genericEntity E
	var i interface{} = genericEntity
	_, ok := i.(config.TimeSeriesIndexeder)
	if ok {
		return d.UpdateTimeSeries(ents)
	}

	kvEnts := make([]sm.Entry, 0)
	for idx, e := range ents {

		proposal, entity, err := d.parseProposal(e.Cmd)
		if err != nil {
			return kvEnts, err
		}

		kvdata := &KVData{
			Key: []byte(strconv.FormatUint(entity.Identifier(), 10)),
			Val: proposal.Data}

		jsonDataKV, err := json.Marshal(kvdata)
		if err != nil {
			return nil, err
		}

		entry := sm.Entry{Index: e.Index, Cmd: jsonDataKV}

		if proposal.Query == QUERY_TYPE_DELETE {
			err = d.diskKV.Delete(entry)
			if err != nil {
				d.logger.Error(err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)
		ents[idx].Result = sm.Result{Value: ents[idx].Index}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}

func (d *GenericDiskKV[E]) GetPageUsingTimeSeriesIndex(pageQuery query.Page) (interface{}, error) {
	var entities = make([]E, 0)

	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageQuery.PageSize

	db := (*pebbledb)(atomic.LoadPointer(&d.diskKV.db))
	iter := db.db.NewIter(db.ro)
	defer iter.Close()

	// Query the database TimeSeries index
	i := 0
	//entityIDs := make([]uint64, 0)
	entityKeys := make([][]byte, 0)
	for iter.First(); iter.Valid(); iter.Next() {

		if i < offset {
			i++
			continue
		}
		if i == offset+pageQuery.PageSize {
			break
		}
		i++

		key := iter.Key()
		value := iter.Value()

		if string(key) == "applied_index" {
			// This is the dragonboat applied_index record
			continue
		}

		timeSeriesIndex := index.NewTimeSeriesIndex()
		if err := timeSeriesIndex.ParseKeyValue(key, value); err != nil {
			if err == index.ErrInvalidKeyPrefix {
				break // Iterator has scrolled past the timeseries record
			}
			return nil, err
		}

		//entityIDs = append(entityIDs, timeSeriesIndex.EntityID)
		entityKeys = append(entityKeys, timeSeriesIndex.EntityIDKey)
	}

	// Retrieve all of the entities from the TimeSeries index result set
	entityValues, err := d.diskKV.LookupBatch(entityKeys)
	if err != nil {
		return nil, err
	}

	// Decode all of the returned entities being returned as []byte
	for _, ev := range entityValues {
		var entity E
		err = json.Unmarshal(ev.([]byte), &entity)
		if err != nil {
			d.logger.Errorf("[Page] Error: %s\n", err)
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func (d *GenericDiskKV[E]) UpdateTimeSeries(ents []sm.Entry) ([]sm.Entry, error) {

	entities := make([]config.TimeSeriesIndexeder, len(ents))
	kvEnts := make([]sm.Entry, 0)
	for idx, e := range ents {

		proposal, entity, err := d.parseProposalWithTimeSeries(e.Cmd)
		if err != nil {
			return kvEnts, err
		}

		entry := sm.Entry{Index: e.Index, Cmd: proposal.Data}

		if proposal.Query == QUERY_TYPE_DELETE {
			err = d.diskKV.DeleteWithTimeSeriesIndex(entry, entity)
			if err != nil {
				d.logger.Error(err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)
		entities[idx] = entity
		ents[idx].Result = sm.Result{Value: ents[idx].Index}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.UpdateWithTimeSeriesIndex(kvEnts, entities)
	}

	return ents, nil
}

func (d *GenericDiskKV[E]) parseProposal(cmd []byte) (Proposal, config.KeyValueEntity, error) {
	var proposal Proposal
	err := json.Unmarshal(cmd, &proposal)
	if err != nil {
		d.logger.Errorf("[entityMachine.Update] Error: %s\n", err)
		return proposal, nil, err
	}

	var entity E
	err = json.Unmarshal(proposal.Data, &entity)
	if err != nil {
		d.logger.Errorf("[entityMachine.Update] Error: %s\n", err)
		return proposal, nil, err
	}

	if _, ok := any(entity).(config.KeyValueEntity); !ok {
		return proposal, nil, errors.New("entity doesn't implement KeyValueEntity interface")
	}
	kvEntity := any(entity).(config.KeyValueEntity)
	return proposal, kvEntity, nil
}

func (d *GenericDiskKV[E]) parseProposalWithTimeSeries(cmd []byte) (Proposal, config.TimeSeriesIndexeder, error) {
	var proposal Proposal
	err := json.Unmarshal(cmd, &proposal)
	if err != nil {
		d.logger.Errorf("[entityMachine.Update] Error: %s\n", err)
		return proposal, nil, err
	}

	var entity E
	err = json.Unmarshal(proposal.Data, &entity)
	if err != nil {
		d.logger.Errorf("[entityMachine.Update] Error: %s\n", err)
		return proposal, nil, err
	}

	if _, ok := any(entity).(config.TimeSeriesIndexeder); !ok {
		return proposal, nil, errors.New("entity doesn't implement TimeSeriesIndexeder interface")
	}
	tsEntity := any(entity).(config.TimeSeriesIndexeder)
	return proposal, tsEntity, nil
}
