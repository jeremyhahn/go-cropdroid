//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"io"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type SystemMachine interface {
	CreateSystemDiskKV(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type SystemDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	SystemMachine
}

func NewSystemDiskKV(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) SystemMachine {

	return &SystemDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *SystemDiskKV) CreateSystemDiskKV(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *SystemDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *SystemDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *SystemDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *SystemDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *SystemDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *SystemDiskKV) Close() error {
	return d.diskKV.Close()
}

// Lookup expects the uint64 organization ID as the key
func (d *SystemDiskKV) Lookup(key interface{}) (interface{}, error) {
	return d.diskKV.Lookup(key)
}

func (d *SystemDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	return d.diskKV.Update(ents)
}
