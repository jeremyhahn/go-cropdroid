// +build !cluster

package model

import (
	"github.com/jeremyhahn/go-cropdroid/app"
)

// https://github.com/google/pprof
// https://golang.org/pkg/net/http/pprof/
// https://golang.org/pkg/runtime

type SystemRuntime struct {
	Version     string `json:"version"`
	Goroutines  int    `json:"goroutines"`
	Cpus        int    `json:"cpus"`
	Cgo         int64  `json:"cgo"`
	HeapSize    uint64 `json:"heapAlloc"`
	Alloc       uint64 `json:"alloc"`
	Sys         uint64 `json:"sys"`
	Mallocs     uint64 `json:"mallocs"`
	Frees       uint64 `json:"frees"`
	NumGC       uint32 `json:"gc"`
	NumForcedGC uint32 `json:"forcedgc"`
}

type System struct {
	Mode                    string          `json:"mode"`
	Version                 *app.AppVersion `json:"version"`
	Farms                   int             `json:"farms"`
	Changefeeds             int             `json:"changefeeds"`
	NotificationQueueLength int             `json:"notificationQueueLength"`
	DeviceIndexLength   int             `json:"deviceIndexLength"`
	ChannelIndexLength      int             `json:"channelIndexLength"`
	Runtime                 *SystemRuntime  `json:"runtime"`
}
