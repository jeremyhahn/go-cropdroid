//go:build cluster && pebble
// +build cluster,pebble

package statemachine

import (
	"errors"
)

const (
	QUERY_TYPE_WILDCARD = iota
	QUERY_TYPE_UPDATE
	QUERY_TYPE_DELETE
	QUERY_TYPE_COUNT
)

var (
	ErrUnsupportedQuery = errors.New("unsupported raft query")
	ErrNullDataProposal = errors.New("null raft proposal data")
)
