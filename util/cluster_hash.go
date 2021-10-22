package util

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/jeremyhahn/go-cropdroid/common"
)

type IdGenerator interface {
	NewID(string) uint64
	Uint64Bytes(uint64) []byte
	StringBytes(str string) []byte
}

type Fnv1aHasher struct {
	is64bit bool
	IdGenerator
}

func NewIdGenerator(datastoreEngine string) IdGenerator {
	uid := &Fnv1aHasher{}
	if datastoreEngine == common.DATASTORE_TYPE_SQLITE ||
		datastoreEngine == common.DATASTORE_TYPE_POSTGRES {
		uid.is64bit = false
	} else {
		uid.is64bit = true
	}
	return uid
}

// Returns a new 64-bit FNV-1a hash from a string
func (hasher *Fnv1aHasher) NewID(str string) uint64 {
	id := hasher.createClusterHash([]byte(str))
	if hasher.is64bit {
		return id
	}
	// Clear Most Significant Bit
	id &^= (1 << 63)
	return id
}

// Returns a new 64-bit FNV-1a hash as little endian byte array
func (hasher *Fnv1aHasher) Uint64Bytes(id uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, id)
	return bytes
}

func (hasher *Fnv1aHasher) StringBytes(str string) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, hasher.NewID(str))
	return bytes
}

// Returns a new 64-bit FNV-1a hash from a byte array
func (hasher *Fnv1aHasher) createClusterHash(bytes []byte) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write(bytes)
	return clusterHash.Sum64()
}
