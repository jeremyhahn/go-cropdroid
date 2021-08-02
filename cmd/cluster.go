// +build cluster,!cloud

package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/cluster"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice"

	"github.com/spf13/cobra"
)

var ClusterID int
var ClusterJoin bool
var ClusterVirtualNodes int
var ClusterRegion string
var ClusterZone string
var ClusterMaxNodes int
var ClusterIaasProvider string
var ClusterListenAddress string
var ClusterGossipPeers string
var ClusterGossipPort int
var ClusterRaft string
var ClusterRaftPort int
var ClusterRaftLeaderID int
var ClusterBootstrap int

func init() {

	// IaaS
	clusterCmd.PersistentFlags().StringVarP(&ClusterIaasProvider, "provider", "", "kvm",
		"Infrastructure-as-a-Service (IaaS) provider [ libvirt | terraform ]")

	// General cluster
	clusterCmd.PersistentFlags().IntVarP(&ClusterID, "cluster-id", "", 0, "Cluster unique identifier")
	clusterCmd.PersistentFlags().IntVarP(&ClusterVirtualNodes, "vnodes", "", 1, "Number of virtual nodes on the consistent hash ring")
	clusterCmd.PersistentFlags().BoolVarP(&ClusterJoin, "join", "", false, "True to join an existing cluster, false to create a new cluster")
	clusterCmd.PersistentFlags().StringVarP(&ClusterListenAddress, "listen", "", "", "The IP address services listen for incoming requests")
	//clusterCmd.PersistentFlags().IntVarP(&ClusterBootstrap, "bootstrap", "", 0, "Number of nodes to wait on when bootstrapping the cluster")
	//clusterCmd.PersistentFlags().StringVarP(&ClusterGossipPeers, "peers", "", "localhost:63001,localhost:63002,localhost:63003", "Cluster member peer addresses")

	// Gossip
	clusterCmd.PersistentFlags().StringVarP(&ClusterGossipPeers, "gossip-peers", "", "", "Comma delimited list of gossip peers (ip:port,ip2:port2)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterGossipPort, "gossip-port", "", 60010, "Gossip server port")
	clusterCmd.PersistentFlags().StringVarP(&ClusterRegion, "region", "", "us-east-1", "The Gossip region this node is placed")
	clusterCmd.PersistentFlags().StringVarP(&ClusterZone, "zone", "", "z1", "The zone within the region which this node is placed")

	// Raft
	clusterCmd.PersistentFlags().IntVarP(&ClusterRaftLeaderID, "raft-leader-id", "", 0, "Starts Raft cluster requesting this node ID become leader")
	clusterCmd.PersistentFlags().StringVarP(&ClusterRaft, "raft", "", "", "Initial Raft members (for bootstrapping a new cluster)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterRaftPort, "raft-port", "", 60020, "Initial Raft members (for bootstrapping a new cluster)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterMaxNodes, "raft-max-nodes", "", 7, "Maximum number of nodes allowed to join a raft cluster")

	clusterCmd.PersistentFlags().StringVarP(&DeviceStore, "device-store", "", "raft", "The database that stores historical device data [ datastore | redis | raft ]")
	DeviceStore = strings.ToLower(DeviceStore)

	rootCmd.AddCommand(clusterCmd)
}

/*
https://github.com/lni/dragonboat/blob/master/CHANGELOG.md
func gprcFactoryFunc(nhc NodeHostConfig, raftio.RequestHandler, raftio.IChunkHandler) raftio.IRaftRPC {
}*/

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Run the cropdroid service in cluster mode",
	Long: `Starts the cropdroid real-time protection and notification service
	using a cluster license. The Raft consenus algorithm is combined with an
	embedded rocksdb database to provide a high performance, highly available,
	fault-tolerant, and massively scalable cloud native architecture.`,
	Run: func(cmd *cobra.Command, args []string) {

		App.Mode = common.MODE_CLUSTER
		App.InitGormDB()

		if ClusterRaft != "" {
			pieces := strings.Split(ClusterRaft, ",")
			ClusterBootstrap = len(pieces) // bootstrapping; wait for all raft cluster to initialize
		}

		localAddress := util.ParseLocalIP()
		if ClusterListenAddress == "" {
			App.Logger.Debugf("Cluster listen address undefined, using %s", localAddress)
			ClusterListenAddress = localAddress
		}

		gossipPeers := make([]string, 0)
		if len(ClusterGossipPeers) > 0 {
			gpeers := strings.Split(ClusterGossipPeers, ",")
			for _, peer := range gpeers {
				gossipPeers = append(gossipPeers, strings.TrimSpace(peer))
			}
		}

		raftPeers := strings.Split(ClusterRaft, ",")
		for i, peer := range raftPeers {
			raftPeers[i] = strings.TrimSpace(peer)
		}

		farmProvisionerChan := make(chan config.FarmConfig, common.BUFFERED_CHANNEL_SIZE)
		farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)
		daoRegistry := gorm.NewGormRegistry(App.Logger, App.GORM)

		params := cluster.NewClusterParams(App.Logger, uint64(ClusterID), uint64(NodeID), ClusterIaasProvider, ClusterRegion,
			ClusterZone, App.DataDir, localAddress, ClusterListenAddress, gossipPeers, raftPeers, ClusterJoin, ClusterGossipPort,
			ClusterRaftPort, ClusterRaftLeaderID, ClusterVirtualNodes, ClusterMaxNodes, ClusterBootstrap, daoRegistry,
			farmProvisionerChan, farmTickerProvisionerChan)

		App.GossipCluster = cluster.NewGossipCluster(params, cluster.NewHashring(ClusterVirtualNodes), daoRegistry.GetFarmDAO())
		App.GossipCluster.Join()
		go App.GossipCluster.Run()

		App.RaftCluster = App.GossipCluster.GetSystemRaft()
		for App.RaftCluster == nil {
			App.Logger.Debug("Waiting to be assigned to a Raft cluster...")
			time.Sleep(1 * time.Second)
			App.RaftCluster = App.GossipCluster.GetSystemRaft()
		}

		//App.ConfigStore = cluster.NewRaftFarmConfigStore(App.Logger, App.RaftCluster)
		//App.FarmStore = cluster.NewRaftFarmStateStore(App.Logger, App.RaftCluster)

		// TODO: Check device hardware and firmware versions at startup, update devices db table
		builder := builder.NewClusterConfigBuilder(App, params, DeviceStore, AppStateTTL, AppStateTick)
		_, serverConfig, serviceRegistry, restServices, err := builder.Build()
		if err != nil {
			App.Logger.Fatal(err)
		}

		// farmProvisioner := provisioner.NewFarmProvisioner(App.Logger, App.NewGormDB(),
		// 	App.Location, datastoreRegistry.GetFarmDAO(), serviceRegistry)

		App.Config = serverConfig.(*config.Server)
		//App.DeviceIndex = deviceIndex
		//App.ChannelIndex = channelIndex

		for _, farmService := range serviceRegistry.GetFarmServices() {
			go farmService.RunCluster()
		}

		if changefeedService := serviceRegistry.GetChangefeedService(); changefeedService != nil {
			changefeedService.Subscribe()
		}

		webserver := webservice.NewWebserver(App, serviceRegistry, restServices, farmTickerProvisionerChan)
		go webserver.Run()
		go webserver.RunClusterProvisionerConsumer()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT) // syscall.SIGTERM, syscall.SIGHUP)

		<-stop

		App.Logger.Info("Shutting down...")

		//webserver.Shutdown()

		if err := App.GossipCluster.Shutdown(); err != nil {
			App.Logger.Error(err)
		}

		if err := App.RaftCluster.Shutdown(); err != nil {
			App.Logger.Fatal(err)
		}

		App.Logger.Info("Shutdown complete")
	},
}
