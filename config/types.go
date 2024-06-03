package config

import (
	"errors"
)

const (
	MEMORY_STORE = iota
	GORM_STORE
	RAFT_MEMORY_STORE
	RAFT_DISK_STORE
)

const (
	CONFIG_KEY_NAME     = "name"
	CONFIG_KEY_MODE     = "mode"
	CONFIG_KEY_INTERVAL = "interval"
	CONFIG_KEY_TIMEZONE = "timezone"
	CONFIG_KEY_ENABLE   = "enable"
	CONFIG_KEY_NOTIFY   = "notify"
	CONFIG_KEY_URI      = "uri"
)

var (
	ErrDeviceNotFound       = errors.New("device not found")
	ErrWorkflowNotFound     = errors.New("workflow not found")
	ErrWorkflowStepNotFound = errors.New("workflow step not found")
)

type KeyValueEntity interface {
	SetID(id uint64)
	Identifier() uint64
}

type TimeSeriesIndexeder interface {
	KeyValueEntity
	SetTimestamp(timestamp uint64)
	Timestamp() uint64
}
