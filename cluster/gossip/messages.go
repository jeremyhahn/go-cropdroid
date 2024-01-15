package gossip

import (
	"fmt"
)

type ClusterEventType int

const (
	EventWorkerAvailable ClusterEventType = iota
	EventWorkerAssigned
	EventProvisionRequest
)

func (t ClusterEventType) String() string {
	switch t {
	case EventWorkerAvailable:
		return "worker-available"
	case EventWorkerAssigned:
		return "worker-assigned"
	case EventProvisionRequest:
		return "provision-request"
	default:
		panic(fmt.Sprintf("unknown event type: %d", t))
	}
}

type WorkerReadyRequest struct {
	Type    string `json:"t"`
	Address string `json:"a"`
}

type WorkerReadyResponse struct {
	Type    string `json:"t"`
	Address string `json:"a"`
}

type WorkerAssignedMessage struct {
	Type          string `json:"t"`
	NodeID        uint64 `json:"n"`
	MemberAddress string `json:"ma"`
	RaftAddress   string `json:"ra"`
}

type ProvisionRequest struct {
	NodeID           uint64 `json:"nid"`
	ConfigClusterID  uint64 `json:"cid"`
	StateClusterID   uint64 `json:"sid"`
	OrganizationID   uint64 `json:"oid"`
	ConfigStoreType  int    `json:"cst"`
	StateStoreType   int    `json:"sst"`
	DataStoreType    int    `json:"dst"`
	ConsistencyLevel int    `json:"cl"`
	FarmName         string `json:"fn"`
	UserID           uint64 `json:"uid"`
	RoleID           uint64 `json:"rid"`
}
