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

type OrganizationConfigMachine interface {
	CreateOrganizationConfigMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine
	sm.IOnDiskStateMachine
}

type OrganizationDiskKV struct {
	logger      *logging.Logger
	idGenerator util.IdGenerator
	dbPath      string
	diskKV      DiskKV
	OrganizationConfigMachine
}

func NewOrganizationConfigMachine(logger *logging.Logger,
	idGenerator util.IdGenerator, dbPath string,
	clusterID, nodeID uint64) OrganizationConfigMachine {

	return &OrganizationDiskKV{
		logger:      logger,
		idGenerator: idGenerator,
		diskKV: DiskKV{
			dbPath:    dbPath,
			clusterID: clusterID,
			nodeID:    nodeID}}
}

func (d *OrganizationDiskKV) CreateOrganizationConfigMachine(
	clusterID, nodeID uint64) sm.IOnDiskStateMachine {

	d.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	d.diskKV.clusterID = clusterID
	d.diskKV.nodeID = nodeID
	return d
}

func (d *OrganizationDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	return d.diskKV.Open(stopc)
}

func (d *OrganizationDiskKV) Sync() error {
	return d.diskKV.Sync()
}

func (d *OrganizationDiskKV) PrepareSnapshot() (interface{}, error) {
	return d.diskKV.PrepareSnapshot()
}

func (d *OrganizationDiskKV) SaveSnapshot(ctx interface{}, w io.Writer, done <-chan struct{}) error {
	return d.diskKV.SaveSnapshot(ctx, w, done)
}

func (d *OrganizationDiskKV) RecoverFromSnapshot(r io.Reader, done <-chan struct{}) error {
	return d.diskKV.RecoverFromSnapshot(r, done)
}

func (d *OrganizationDiskKV) Close() error {
	return d.diskKV.Close()
}

// Lookup expects the uint64 organization ID as the key
func (d *OrganizationDiskKV) Lookup(key interface{}) (interface{}, error) {

	switch t := key.(type) {

	case uint64:
		id := d.idGenerator.Uint64Bytes(t)
		v, err := d.diskKV.Lookup(id)
		if err != nil {
			return nil, err
		}
		var organizationConfig config.Organization
		err = json.Unmarshal(v.([]byte), &organizationConfig)
		if err != nil {
			d.logger.Errorf("[OrganizationDiskKV.Lookup] Error: %s\n", err)
			return nil, err
		}
		return &organizationConfig, err

	case []uint8:
		query, err := strconv.Atoi(fmt.Sprintf("%s", key))
		if err != nil {
			d.logger.Errorf("[OrganizationDiskKV.Lookup] Error: %s\n", err)
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
			orgs := make([]*config.Organization, 0)
			for _, orgKV := range values {
				if string(orgKV.Key) == appliedIndexKey {
					continue
				}
				var org config.Organization
				if err := json.Unmarshal(orgKV.Val, &org); err != nil {
					return nil, err
				}
				for _, farm := range org.GetFarms() {
					farm.ParseSettings()
					//org.Farms[i] = farm
				}
				orgs = append(orgs, &org)
			}
			return orgs, nil
		}
		return nil, nil
	}
	return nil, datastore.ErrNotFound
}

// Updates and deletes items from the ondisk state machine
func (d *OrganizationDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	kvEnts := make([]sm.Entry, 0)

	for idx, e := range ents {

		var proposal Proposal
		err := json.Unmarshal(e.Cmd, &proposal)
		if err != nil {
			d.logger.Errorf("[OrganizationConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		var organizationConfig config.Organization
		err = json.Unmarshal(proposal.Data, &organizationConfig)
		if err != nil {
			d.logger.Errorf("[OrganizationConfigMachine.Update] Error: %s\n", err)
			return nil, err
		}

		kvdata := &KVData{
			//Key: []byte(fmt.Sprint(organizationConfig.ID)),
			Key: d.idGenerator.Uint64Bytes(organizationConfig.ID),
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
				d.logger.Errorf("[OrganizationConfigMachine.Update] Error: %s\n", err)
				return nil, err
			}
			continue
		}

		kvEnts = append(kvEnts, entry)

		//d.organizationConfigChangeChan <- &organizationConfig
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}

	if len(kvEnts) > 0 {
		return d.diskKV.Update(kvEnts)
	}

	return ents, nil
}
