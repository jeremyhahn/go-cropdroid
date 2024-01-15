//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/test/data"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/lni/dragonboat/v3/statemachine"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationOpen(t *testing.T) {

	defer cleanup(databasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)

	//farmChangeConfigChan := make(chan config.FarmConfig, 5)
	orgConfigMachine := NewOrganizationConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	lastAppliedIndex, err := orgConfigMachine.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)
}

func TestOrganizationUpdateLookupEmptyStore(t *testing.T) {

	defer cleanup(databasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	//farmChangeConfigChan := make(chan config.FarmConfig, 5)
	sm := NewOrganizationConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	org := data.CreateTestOrganization1(idGenerator)

	orgJson, err := json.Marshal(org)
	assert.Nil(t, err)

	lastAppliedIndex, err := sm.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)

	proposal, err := CreateProposal(
		QUERY_TYPE_UPDATE, orgJson).Serialize()

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   proposal}}

	ents, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ents))
	// for i, ent := range ents {
	// 	assert.Equal(t, ent.Result.Value, uint64(len(entries[i].Cmd)))
	// }

	persisted, err := sm.Lookup(org.GetID())
	assert.Nil(t, err)
	assert.NotNil(t, persisted)

	orgConfig := persisted.(*config.Organization)
	assert.True(t, data.OrgsEqual(orgConfig, org))
}

func TestOrganizationPrepareSnapshot(t *testing.T) {

	defer cleanup(databasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	sm := NewOrganizationConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	org := data.CreateTestOrganization1(idGenerator)

	bytes, err := json.Marshal(org)

	assert.Nil(t, err)
	proposal, err := CreateProposal(
		QUERY_TYPE_UPDATE, bytes).Serialize()

	entries := []statemachine.Entry{{
		Index: 1,
		Cmd:   proposal}}

	stopc := make(chan struct{}, 1)
	appliedIndex, err := sm.Open(stopc)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), appliedIndex)

	results, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	//assert.Equal(t, bytes, results[0].Cmd)

	diskKV1, err := sm.PrepareSnapshot()
	assert.Nil(t, err)
	diskKVCtx1 := diskKV1.(*diskKVCtx)
	assert.IsType(t, &pebble.Snapshot{}, diskKVCtx1.snapshot)

	diskKV2, err := sm.PrepareSnapshot()
	assert.Nil(t, err)
	diskKVCtx2 := diskKV2.(*diskKVCtx)
	assert.IsType(t, &pebble.Snapshot{}, diskKVCtx2.snapshot)

	assert.NotEqual(t, diskKV1, diskKV2)
	assert.NotEqual(t, diskKVCtx1, diskKVCtx2)
	//assert.NotEqual(t, diskKVCtx1.db, diskKVCtx2.db)
	assert.NotEqual(t, diskKVCtx1.snapshot, diskKVCtx2.snapshot)
}

func TestOrganizationSaveAndRecoverSnapshot(t *testing.T) {

	defer cleanup(databasePath)

	logger := createLogger()
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	sm := NewOrganizationConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)

	org1 := data.CreateTestOrganization1(idGenerator)
	org2 := data.CreateTestOrganization2(idGenerator)

	org1Bytes, err := json.Marshal(org1)
	assert.Nil(t, err)

	org2Bytes, err := json.Marshal(org2)
	assert.Nil(t, err)

	proposal1, err := CreateProposal(
		QUERY_TYPE_UPDATE, org1Bytes).Serialize()

	proposal2, err := CreateProposal(
		QUERY_TYPE_UPDATE, org2Bytes).Serialize()

	entries := []statemachine.Entry{
		{
			Index: 1,
			Cmd:   proposal1},
		{
			Index: 2,
			Cmd:   proposal2}}

	lastAppliedIndex, err := sm.Open(nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, uint64(0), lastAppliedIndex)

	results, err := sm.Update(entries)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results))
	// Cmd gets wrapped in dataKV so this compare fails
	//assert.Equal(t, org1Bytes, results[0].Cmd)
	//assert.Equal(t, org2Bytes, results[1].Cmd)

	diskKV1, err := sm.PrepareSnapshot()
	assert.Nil(t, err)
	diskKVCtx1 := diskKV1.(*diskKVCtx)
	assert.IsType(t, &pebble.Snapshot{}, diskKVCtx1.snapshot)

	done := make(chan struct{}, 1)

	var snapshot bytes.Buffer

	err = sm.SaveSnapshot(diskKV1, &snapshot, done)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(done))

	sm2 := NewOrganizationConfigMachine(logger,
		idGenerator, databasePath, clusterID, nodeID)
	err = sm2.RecoverFromSnapshot(&snapshot, done)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(done))

	result1, err := sm2.Lookup(org1.GetID())
	assert.Nil(t, err)
	assert.NotNil(t, result1)

	persistedOrg1 := result1.(*config.Organization)
	assert.True(t, data.OrgsEqual(org1, persistedOrg1))

	result2, err := sm2.Lookup(org2.GetID())
	assert.Nil(t, err)
	assert.NotNil(t, result2)

	persistedOrg2 := result2.(*config.Organization)
	assert.True(t, data.OrgsEqual(org2, persistedOrg2))
}
