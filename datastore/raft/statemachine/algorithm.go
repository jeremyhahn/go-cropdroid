//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"io"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
	sm "github.com/lni/dragonboat/v3/statemachine"
	logging "github.com/op/go-logging"
)

type AlgorithmDiskKV struct {
	GenericDiskKV[*config.Algorithm]
	OnDiskStateMachine
}

func NewAlgorithmConfigMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) OnDiskStateMachine {

	return &AlgorithmDiskKV{
		GenericDiskKV: GenericDiskKV[*config.Algorithm]{
			logger:      logger,
			idGenerator: idGenerator,
			diskKV: DiskKV[*config.Algorithm]{
				idGenerator: idGenerator,
				dbPath:      dbPath,
				clusterID:   clusterID,
				nodeID:      nodeID}}}
}

func (d *AlgorithmDiskKV) CreateOnDiskStateMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	return d.GenericDiskKV.CreateOnDiskStateMachine(clusterID, nodeID)
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
	return d.GenericDiskKV.Lookup(key)
}

func (d *AlgorithmDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	return d.GenericDiskKV.Update(ents)
}
