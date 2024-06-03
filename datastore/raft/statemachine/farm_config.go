//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"io"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type FarmConfigOnDiskStateMachine interface {
	CreateFarmConfigOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
}

type FarmConfigDiskKV struct {
	GenericDiskKV        GenericDiskKV[*config.FarmStruct]
	farmConfigChangeChan chan config.Farm
	FarmConfigOnDiskStateMachine
}

func NewFarmConfigOnDiskStateMachine(logger *logging.Logger, idGenerator util.IdGenerator,
	dbPath string, clusterID, nodeID uint64, farmConfigChangeChan chan config.Farm) FarmConfigOnDiskStateMachine {

	return &FarmConfigDiskKV{
		GenericDiskKV: GenericDiskKV[*config.FarmStruct]{
			logger:      logger,
			idGenerator: idGenerator,
			diskKV: DiskKV[*config.FarmStruct]{
				idGenerator: idGenerator,
				dbPath:      dbPath,
				clusterID:   clusterID,
				nodeID:      nodeID}}}
}

func (d *FarmConfigDiskKV) CreateFarmConfigOnDiskStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine {
	d.GenericDiskKV.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.GenericDiskKV.diskKV.clusterID = clusterID
	d.GenericDiskKV.diskKV.nodeID = nodeID
	return d
}

func (d *FarmConfigDiskKV) Close() error {
	return d.GenericDiskKV.Close()
}

func (d *FarmConfigDiskKV) Count() int64 {
	return d.GenericDiskKV.Count()
}

func (d *FarmConfigDiskKV) GetPage(pageQuery query.PageQuery) (interface{}, error) {
	return d.GenericDiskKV.GetPage(pageQuery)
}

func (d *FarmConfigDiskKV) GetPageUsingTimeSeriesIndex(pageQuery query.PageQuery) (interface{}, error) {
	return d.GenericDiskKV.GetPageUsingTimeSeriesIndex(pageQuery)
}

func (d *FarmConfigDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.GenericDiskKV.Open(stopc)
}

func (d *FarmConfigDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.GenericDiskKV.PrepareSnapshot()
}

func (d *FarmConfigDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.GenericDiskKV.RecoverFromSnapshot(r, done)
}

func (d *FarmConfigDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.GenericDiskKV.SaveSnapshot(ctx, w, done)
}

func (d *FarmConfigDiskKV) Sync() error {
	return d.GenericDiskKV.Sync()
}

func (d *FarmConfigDiskKV) Lookup(key interface{}) (interface{}, error) {
	return d.GenericDiskKV.Lookup(key)
}

func (d *FarmConfigDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	return d.GenericDiskKV.Update(ents)
}

func (d *FarmConfigDiskKV) UpdateTimeSeries(ents []sm.Entry) ([]sm.Entry, error) {
	return d.GenericDiskKV.UpdateTimeSeries(ents)
}
