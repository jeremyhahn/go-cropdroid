package util

import (
	"fmt"
	"hash/fnv"
)

// Returns a new 64-bit FNV-1a hash
func ClusterHash(key1, key2 uint64) uint64 {
	clusterHash := fnv.New64a()
	clusterHash.Write([]byte(fmt.Sprintf("%d-%d", key1, key2)))
	return clusterHash.Sum64()
}
