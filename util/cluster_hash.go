package util

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

// Returns a new 64-bit FNV-1a hash

func NewClusterHash(key1, key2 uint64) uint64 {
	return CreateClusterHash([]byte(fmt.Sprintf("%d-%d", key1, key2)))
}

// Returns a new 64-bit FNV-1a hash from a byte array
func CreateClusterHash(bytes []byte) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write(bytes)
	return clusterHash.Sum64()
}

func ClusterHashFromString(str string) uint64 {
	return CreateClusterHash([]byte(str))
}

// Returns a new 64-bit FNV-1a hash as little endian byte array
func ClusterHashAsBytes(key1, key2 uint64) []byte {
	chash := NewClusterHash(key1, key2)
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

func NewOrganizationHash(orgName string) uint64 {
	return CreateClusterHash([]byte(orgName))
}

func NewRegistrationHash(key1 string, key2 int64) uint64 {
	key := fmt.Sprintf("%s-%d", key1, uint64(10))
	return CreateClusterHash([]byte(key))
}
