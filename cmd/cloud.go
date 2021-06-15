// +build cloud

package cmd

import (
	"strings"

	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/webservice"

	"github.com/spf13/cobra"
)

var CloudID int
var CloudPeers string
var CloudJoin bool
var CloudDevFlag bool
var CloudGossipPort int
var GoogleCredentialsFile string

func init() {

	cloudCmd.PersistentFlags().IntVarP(&CloudID, "cloud-id", "", 0, "Cloud unique identifier")
	cloudCmd.PersistentFlags().StringVarP(&CloudPeers, "peers", "", "localhost:64001,localhost:64002,localhost:64003", "Cloud member peer addresses")
	cloudCmd.PersistentFlags().BoolVarP(&CloudJoin, "join", "", false, "True to join an existing cloud, false to create a new cloud")
	cloudCmd.PersistentFlags().BoolVarP(&CloudDevFlag, "dev", "", false, "True to create a local development cloud")

	// Gossip
	cloudCmd.PersistentFlags().IntVarP(&CloudGossipPort, "gossip-port", "", 8946, "Gossip port communication")

	rootCmd.AddCommand(cloudCmd)
}

/*
https://github.com/lni/dragonboat/blob/master/CHANGELOG.md
func gprcFactoryFunc(nhc NodeHostConfig, raftio.RequestHandler, raftio.IChunkHandler) raftio.IRaftRPC {
}*/

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Run the cropdroid service in cloud mode",
	Long: `Starts the cropdroid real-time protection and notification service
	using a cloud license. The Raft consenus algorithm is combined with an
	embedded rocksdb database to provide a high performance, highly available,
	fault-tolerant, and massively scalable cloud native architecture.`,
	Run: func(cmd *cobra.Command, args []string) {

		// As a summary, when -
		//  - starting a brand new Raft cloud, set join to false and specify all initial
		//    member node details in the initialMembers map.
		//  - joining a new node to an existing Raft cloud, set join to true and leave
		//    the initialMembers map empty. This requires the joining node to have already
		//    been added as a member node of the Raft cloud.
		//  - restarting an crashed or stopped node, set join to false and leave the
		//    initialMembers map to be empty. This applies to both initial member nodes
		//    and those joined later.

		peers := strings.Split(CloudPeers, ",")
		for i, peer := range peers {
			peers[i] = strings.TrimSpace(peer)
		}

		params := cluster.NewClusterParams(App.Logger, uint64(CloudID), uint64(NodeID), App.DataDir, peers, CloudJoin, CloudGossipPort)

		App.EnableRegistrations = true
		App.RaftCluster = cluster.NewRaftCluster(params)
		//App.FarmStore = cloud.NewRaftFarmStateStore(App.Logger, App.RaftCloud)

		App.InitGormDB()

		if MetricDatastore == "raft" {
			App.MetricDatastore = cluster.NewRaftDeviceStateDAO(App.Logger, App.RaftCluster)
		} else if MetricDatastore == "datastore" {
			// Don't subscribe to device state changefeeds while running Raft cluster!
			App.MetricDatastore = gorm.NewDeviceStateDAO(App.Logger, App.GORM, App.GORMInitParams.Engine, App.Location)
		} else if MetricDatastore == "redis" {
			App.MetricDatastore = datastore.NewRedisDeviceStateDAO(":6379", "")
		}

		// TODO: Check device hardware and firmware versions at startup, update devices db table
		serverConfig, serviceRegistry, restServices, deviceIndex, channelIndex, err := builder.NewCloudConfigBuilder(App, params).Build()
		if err != nil {
			App.Logger.Fatal(err)
		}

		App.Config = serverConfig.(*config.Server)
		App.DeviceIndex = deviceIndex
		App.ChannelIndex = channelIndex

		go cluster.NewGossipCluster(params)

		webserver := webservice.NewWebserver(App, serviceRegistry, restServices)
		go webserver.Run()

		stop := make(chan bool, 1)
		<-stop

	},
}
