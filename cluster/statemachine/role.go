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

type RoleConfigMachine interface {
	CreateRoleConfigMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type RoleDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	RoleConfigMachine
}

func NewRoleConfigMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) RoleConfigMachine {

	return &RoleDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *RoleDiskKV) CreateRoleConfigMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *RoleDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *RoleDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *RoleDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *RoleDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *RoleDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *RoleDiskKV) Close() error {
	return d.diskKV.Close()
}

func (d *RoleDiskKV) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var roleConfig config.Role
		err = json.Unmarshal(v.([]byte), &roleConfig)
		if err != nil {
			d.logger.Errorf("[RoleDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &roleConfig, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[RoleDiskKV.Lookup] Error: %s\n", err)
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
			roles := make([]*config.Role, 0)
			for _, roleKV := range values {
				if string(roleKV.Key) == appliedIndexKey {
					continue
				}
				var role config.Role
				if err := json.Unmarshal(roleKV.Val, &role); err != nil {
					return nil, err
				}
				roles = append(roles, &role)
			}
			return roles, nil
		}
		return nil, nil
	}
	return nil, datastore.ErrNotFound
}

func (d *RoleDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[RoleConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var roleConfig config.Role
		err = json.Unmarshal(proposal.Data, &roleConfig)
		if err != nil {
			d.logger.Errorf("[RoleConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			Key: d.idGenerator.Uint64Bytes(roleConfig.GetID()),
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
				d.logger.Errorf("[RoleConfigMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		//d.roleConfigChangeChan <- &roleConfig
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
