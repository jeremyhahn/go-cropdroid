// +build cluster

package service

import (
	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/mapper"
)

type DefaultClusterRegistry struct {
	gossipCluster cluster.GossipCluster
	raftCluster   cluster.RaftCluster
	DefaultServiceRegistry
}

type ClusterServiceRegistry interface {
	GetGossipCluster() cluster.GossipCluster
	GetRaftCluster() cluster.RaftCluster
	ServiceRegistry
}

func CreateClusterServiceRegistry(_app *app.App, daos datastore.DatastoreRegistry,
	mappers mapper.MapperRegistry, gossipCluster cluster.GossipCluster,
	raftCluster cluster.RaftCluster) ClusterServiceRegistry {

	registry := CreateServiceRegistry(_app, daos, mappers)
	return &DefaultClusterRegistry{
		DefaultServiceRegistry: *registry.(*DefaultServiceRegistry),
		gossipCluster:          gossipCluster,
		raftCluster:            raftCluster}
}

func (clusterRegistry *DefaultClusterRegistry) GetGossipCluster() cluster.GossipCluster {
	return clusterRegistry.gossipCluster
}

func (clusterRegistry *DefaultClusterRegistry) GetRaftCluster() cluster.RaftCluster {
	return clusterRegistry.raftCluster
}
