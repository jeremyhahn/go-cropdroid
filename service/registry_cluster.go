//go:build cluster
// +build cluster

package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type DefaultClusterRegistry struct {
	daos       dao.Registry
	gossipNode cluster.GossipNode
	raftNode   cluster.RaftNode
	DefaultServiceRegistry
}

type ClusterServiceRegistry interface {
	GetGossipNode() cluster.GossipNode
	GetRaftNode() cluster.RaftNode
	ServiceRegistry
}

func CreateClusterServiceRegistry(_app *app.App, daos dao.Registry,
	mappers mapper.MapperRegistry, gossipNode cluster.GossipNode,
	raftNode cluster.RaftNode) ClusterServiceRegistry {

	registry := CreateServiceRegistry(_app, daos, mappers)
	return &DefaultClusterRegistry{
		DefaultServiceRegistry: *registry.(*DefaultServiceRegistry),
		gossipNode:             gossipNode,
		raftNode:               raftNode}
}

func (clusterRegistry *DefaultClusterRegistry) GetGossipNode() cluster.GossipNode {
	return clusterRegistry.gossipNode
}

func (clusterRegistry *DefaultClusterRegistry) GetRaftNode() cluster.RaftNode {
	return clusterRegistry.raftNode
}
