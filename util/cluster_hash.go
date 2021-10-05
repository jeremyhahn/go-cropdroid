package util

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

// Returns a new 64-bit FNV-1a hash
func ClusterHash(key1, key2 uint64) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", key1, key2)))
	return clusterHash.Sum64()
}

// Returns a new 64-bit FNV-1a hash as little endian byte array
func ClusterHashBytes(key1, key2 uint64) []byte {
	chash := ClusterHash(key1, key2)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, chash)
	return bytes
}

// Returns a new 64-bit FNV-1a hash as little endian byte array
func ClusterIdBytes(clusterID uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, clusterID)
	return bytes
}
