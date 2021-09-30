// +build notnow,!cluster,!pebble

package state

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/jeremyhahn/go-cropdroid/config"
	sm "github.com/lni/dragonboat/v3/statemachine"
	"github.com/tecbot/gorocksdb"
)

// FarmDiskKV is a state machine that implements the IOnDiskStateMachine interface.
// FarmDiskKV stores key-value pairs in the underlying RocksDB key-value store.
type FarmDiskKV struct {
	clusterID            uint64
	nodeID               uint64
	farmConfigChangeChan chan config.FarmConfig
	lastApplied          uint64
	db                   unsafe.Pointer
	closed               bool
	aborted              bool
	dataDir              string
}

// NewFarmDiskKV creates a new disk backed farm config state machine
func NewFarmDiskKV(dataDir string, farmConfigChangeChan chan config.FarmConfig) *FarmDiskKV {
	return &FarmDiskKV{
		dataDir:              dataDir,
		farmConfigChangeChan: farmConfigChangeChan}
}

func (d *FarmDiskKV) CreateStateMachine(clusterID, nodeID uint64) sm.IOnDiskStateMachine {
	d.clusterID = clusterID
	d.nodeID = nodeID
	return d
}

func CreateFarmDiskKV(clusterID uint64, nodeID uint64) sm.IOnDiskStateMachine {
	return &FarmDiskKV{
		clusterID: clusterID,
		nodeID:    nodeID,
	}
}

func (d *FarmDiskKV) queryAppliedIndex(db *rocksdb) (uint64, error) {
	val, err := db.db.Get(db.ro, []byte(appliedIndexKey))
	if err != nil {
		return 0, err
	}
	defer val.Free()
	data := val.Data()
	if len(data) == 0 {
		return 0, nil
	}
	return strconv.ParseUint(string(data), 10, 64)
}

// Open opens the state machine and return the index of the last Raft Log entry
// already updated into the state machine.
func (d *FarmDiskKV) Open(stopc <-chan struct{}) (uint64, error) {
	dir := getNodeDBDirName(d.clusterID, d.nodeID, d.dataDir)
	if err := createNodeDataDir(dir); err != nil {
		panic(err)
	}
	var dbdir string
	if !isNewRun(dir) {
		if err := cleanupNodeDataDir(dir); err != nil {
			return 0, err
		}
		var err error
		dbdir, err = getCurrentDBDirName(dir)
		if err != nil {
			return 0, err
		}
		if _, err := os.Stat(dbdir); err != nil {
			if os.IsNotExist(err) {
				panic("db dir unexpectedly deleted")
			}
		}
	} else {
		dbdir = getNewRandomDBDirName(dir)
		if err := saveCurrentDBDirName(dir, dbdir); err != nil {
			return 0, err
		}
		if err := replaceCurrentDBFile(dir); err != nil {
			return 0, err
		}
	}
	db, err := createDB(dbdir)
	if err != nil {
		return 0, err
	}
	atomic.SwapPointer(&d.db, unsafe.Pointer(db))
	appliedIndex, err := d.queryAppliedIndex(db)
	if err != nil {
		panic(err)
	}
	d.lastApplied = appliedIndex
	return appliedIndex, nil
}

// Lookup queries the state machine.
func (d *FarmDiskKV) Lookup(key interface{}) (interface{}, error) {
	db := (*rocksdb)(atomic.LoadPointer(&d.db))
	if db != nil {
		v, err := db.lookup(key.([]byte))
		if err == nil && d.closed {
			panic("lookup returned valid result when FarmDiskKV is already closed")
		}
		return v, err
	}
	return nil, errors.New("db closed")
}

// Update updates the state machine. In this example, all updates are put into
// a RocksDB write batch and then atomically written to the DB together with
// the index of the last Raft Log entry. For simplicity, we always Sync the
// writes (db.wo.Sync=True). To get higher throughput, you can implement the
// Sync() method below and choose not to synchronize for every Update(). Sync()
// will periodically called by Dragonboat to synchronize the state.
func (d *FarmDiskKV) Update(ents []sm.Entry) ([]sm.Entry, error) {
	if d.aborted {
		panic("update() called after abort set to true")
	}
	if d.closed {
		panic("update called after Close()")
	}
	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()
	db := (*rocksdb)(atomic.LoadPointer(&d.db))
	for idx, e := range ents {
		dataKV := &KVData{}
		if err := json.Unmarshal(e.Cmd, dataKV); err != nil {
			panic(err)
		}
		wb.Put([]byte(dataKV.Key), []byte(dataKV.Val))
		ents[idx].Result = sm.Result{Value: uint64(len(ents[idx].Cmd))}
	}
	// save the applied index to the DB.
	idx := fmt.Sprintf("%d", ents[len(ents)-1].Index)
	wb.Put([]byte(appliedIndexKey), []byte(idx))
	if err := db.db.Write(db.wo, wb); err != nil {
		return nil, err
	}
	if d.lastApplied >= ents[len(ents)-1].Index {
		panic("lastApplied not moving forward")
	}
	d.lastApplied = ents[len(ents)-1].Index
	return ents, nil
}

// Sync synchronizes all in-core state of the state machine. Since the Update
// method in this example already does that every time when it is invoked, the
// Sync method here is a NoOP.
func (d *FarmDiskKV) Sync() error {
	return nil
}

type diskKVCtx struct {
	db       *rocksdb
	snapshot *gorocksdb.Snapshot
}

// PrepareSnapshot prepares snapshotting. PrepareSnapshot is responsible to
// capture a state identifier that identifies a point in time state of the
// underlying data. In this example, we use RocksDB's snapshot feature to
// achieve that.
func (d *FarmDiskKV) PrepareSnapshot() (interface{}, error) {
	if d.closed {
		panic("prepare snapshot called after Close()")
	}
	if d.aborted {
		panic("prepare snapshot called after abort")
	}
	db := (*rocksdb)(atomic.LoadPointer(&d.db))
	return &diskKVCtx{
		db:       db,
		snapshot: db.db.NewSnapshot(),
	}, nil
}

// saveToWriter saves all existing key-value pairs to the provided writer.
// As an example, we use the most straight forward way to implement this.
func (d *FarmDiskKV) saveToWriter(db *rocksdb,
	ss *gorocksdb.Snapshot, w io.Writer) error {
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetSnapshot(ss)
	iter := db.db.NewIterator(ro)
	defer iter.Close()
	count := uint64(0)
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		count++
	}
	sz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sz, count)
	if _, err := w.Write(sz); err != nil {
		return err
	}
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := iter.Key()
		val := iter.Value()
		dataKv := &KVData{
			Key: string(key.Data()),
			Val: string(val.Data()),
		}
		data, err := json.Marshal(dataKv)
		if err != nil {
			panic(err)
		}
		binary.LittleEndian.PutUint64(sz, uint64(len(data)))
		if _, err := w.Write(sz); err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

// SaveSnapshot saves the state machine state identified by the state
// identifier provided by the input ctx parameter. Note that SaveSnapshot
// is not suppose to save the latest state.
func (d *FarmDiskKV) SaveSnapshot(ctx interface{},
	w io.Writer, done <-chan struct{}) error {
	if d.closed {
		panic("prepare snapshot called after Close()")
	}
	if d.aborted {
		panic("prepare snapshot called after abort")
	}
	ctxdata := ctx.(*diskKVCtx)
	db := ctxdata.db
	db.mu.RLock()
	defer db.mu.RUnlock()
	ss := ctxdata.snapshot
	return d.saveToWriter(db, ss, w)
}

// RecoverFromSnapshot recovers the state machine state from snapshot. The
// snapshot is recovered into a new DB first and then atomically swapped with
// the existing DB to complete the recovery.
func (d *FarmDiskKV) RecoverFromSnapshot(r io.Reader,
	done <-chan struct{}) error {
	if d.closed {
		panic("recover from snapshot called after Close()")
	}
	dir := getNodeDBDirName(d.clusterID, d.nodeID, d.dataDir)
	dbdir := getNewRandomDBDirName(dir)
	oldDirName, err := getCurrentDBDirName(dir)
	if err != nil {
		return err
	}
	db, err := createDB(dbdir)
	if err != nil {
		return err
	}
	sz := make([]byte, 8)
	if _, err := io.ReadFull(r, sz); err != nil {
		return err
	}
	total := binary.LittleEndian.Uint64(sz)
	wb := gorocksdb.NewWriteBatch()
	for i := uint64(0); i < total; i++ {
		if _, err := io.ReadFull(r, sz); err != nil {
			return err
		}
		toRead := binary.LittleEndian.Uint64(sz)
		data := make([]byte, toRead)
		if _, err := io.ReadFull(r, data); err != nil {
			return err
		}
		dataKv := &KVData{}
		if err := json.Unmarshal(data, dataKv); err != nil {
			panic(err)
		}
		wb.Put([]byte(dataKv.Key), []byte(dataKv.Val))
	}
	if err := db.db.Write(db.wo, wb); err != nil {
		return err
	}
	if err := saveCurrentDBDirName(dir, dbdir); err != nil {
		return err
	}
	if err := replaceCurrentDBFile(dir); err != nil {
		return err
	}
	newLastApplied, err := d.queryAppliedIndex(db)
	if err != nil {
		panic(err)
	}
	if d.lastApplied > newLastApplied {
		panic("last applied not moving forward")
	}
	d.lastApplied = newLastApplied
	old := (*rocksdb)(atomic.SwapPointer(&d.db, unsafe.Pointer(db)))
	if old != nil {
		old.close()
	}
	return os.RemoveAll(oldDirName)
}

// Close closes the state machine.
func (d *FarmDiskKV) Close() error {
	db := (*rocksdb)(atomic.SwapPointer(&d.db, unsafe.Pointer(nil)))
	if db != nil {
		d.closed = true
		db.close()
	} else {
		if d.closed {
			panic("close called twice")
		}
	}
	return nil
}

// GetHash returns a hash value representing the state of the state machine.
func (d *FarmDiskKV) GetHash() (uint64, error) {
	h := md5.New()
	db := (*rocksdb)(atomic.LoadPointer(&d.db))
	ss := db.db.NewSnapshot()
	db.mu.RLock()
	defer db.mu.RUnlock()
	if err := d.saveToWriter(db, ss, h); err != nil {
		return 0, err
	}
	md5sum := h.Sum(nil)
	return binary.LittleEndian.Uint64(md5sum[:8]), nil
}
