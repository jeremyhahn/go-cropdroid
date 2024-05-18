//go:build cluster && pebble
// +build cluster,pebble

package statemachine

const (
	databasePath = "pebble-testdb"
	clusterID    = uint64(1234567890)
	nodeID       = 1
)
