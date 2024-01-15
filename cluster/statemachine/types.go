//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"errors"
	"os"

	logging "github.com/op/go-logging"
)

const (
	QUERY_TYPE_WILDCARD = iota
	QUERY_TYPE_UPDATE
	QUERY_TYPE_DELETE
)

var (
	ErrUnsupportedQuery = errors.New("unsupported raft query")
)

// Tests only

var TestSuiteName = "cropdroid_cluster_farm_config_test"

func createLogger() *logging.Logger {

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	return logger
}

func cleanup(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}
