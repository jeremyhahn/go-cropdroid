//go:build cluster && !cloud
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
	clusterutil "github.com/jeremyhahn/go-cropdroid/cluster/util"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jeremyhahn/go-cropdroid/webservice"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ClusterID            uint64
	ClusterJoin          bool
	ClusterVirtualNodes  int
	ClusterRegion        string
	ClusterZone          string
	ClusterMaxNodes      int
	ClusterIaasProvider  string
	ClusterListenAddress string
	ClusterGossipPeers   string
	ClusterGossipPort    int
	ClusterRaft          string
	ClusterRaftPort      int
	ClusterRaftLeaderID  uint64
	ClusterBootstrap     int

	DefaultConsistencyLevel int
	DefaultConfigStoreType  int
	DefaultStateStoreType   int
	DefaultDataStoreType    int

	RaftOrganizationClusterID uint64
	RaftUserClusterID         uint64
	RaftRoleClusterID         uint64
	RaftAlgorithmClusterID    uint64
	RaftRegistrationClusterID uint64
)

func init() {

	// IaaS
	clusterCmd.PersistentFlags().StringVarP(&ClusterIaasProvider, "provider", "", "kvm",
		"Infrastructure-as-a-Service (IaaS) provider [ libvirt | terraform ]")

	// General cluster
	clusterCmd.PersistentFlags().Uint64VarP(&ClusterID, "cluster-id", "", 1, "Cluster unique identifier")
	clusterCmd.PersistentFlags().IntVarP(&ClusterVirtualNodes, "vnodes", "", 1, "Number of virtual nodes on the consistent hash ring")
	clusterCmd.PersistentFlags().BoolVarP(&ClusterJoin, "join", "", false, "True to join an existing cluster, false to create a new cluster")
	clusterCmd.PersistentFlags().StringVarP(&ClusterListenAddress, "listen", "", "", "The IP address services listen for incoming requests")
	//clusterCmd.PersistentFlags().BoolVarP(&ClusterBootstrap, "bootstrap", "", 0, "Number of nodes to wait on when bootstrapping the cluster")
	//clusterCmd.PersistentFlags().StringVarP(&ClusterGossipPeers, "peers", "", "localhost:63001,localhost:63002,localhost:63003", "Cluster member peer addresses")

	// Gossip
	clusterCmd.PersistentFlags().StringVarP(&ClusterGossipPeers, "gossip-peers", "", "", "Comma delimited list of gossip peers (ip:port,ip2:port2)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterGossipPort, "gossip-port", "", 60010, "Gossip server port")
	clusterCmd.PersistentFlags().StringVarP(&ClusterRegion, "region", "", "us-east-1", "The Gossip region this node is placed")
	clusterCmd.PersistentFlags().StringVarP(&ClusterZone, "zone", "", "z1", "The zone within the region which this node is placed")

	// Raft
	clusterCmd.PersistentFlags().Uint64VarP(&ClusterRaftLeaderID, "raft-leader-id", "", 0, "Starts Raft cluster requesting this node ID become leader")
	clusterCmd.PersistentFlags().StringVarP(&ClusterRaft, "raft", "", "", "Initial Raft members (for bootstrapping a new cluster)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterRaftPort, "raft-port", "", 60020, "Initial Raft members (for bootstrapping a new cluster)")
	clusterCmd.PersistentFlags().IntVarP(&ClusterMaxNodes, "raft-max-nodes", "", 7, "Maximum number of nodes allowed to join a raft cluster")

	// Raft cluster ids
	clusterCmd.PersistentFlags().Uint64VarP(&RaftOrganizationClusterID, "raft-organization-cluster-id", "", 100, "The raft cluster id assigned to the organization database")
	clusterCmd.PersistentFlags().Uint64VarP(&RaftUserClusterID, "raft-user-cluster-id", "", 101, "The raft cluster id assigned to the user database")
	clusterCmd.PersistentFlags().Uint64VarP(&RaftRoleClusterID, "raft-role-cluster-id", "", 102, "The raft cluster id assigned to the role database")
	clusterCmd.PersistentFlags().Uint64VarP(&RaftAlgorithmClusterID, "raft-algorithm-cluster-id", "", 103, "The raft cluster id assigned to the algorithm database")
	clusterCmd.PersistentFlags().Uint64VarP(&RaftRegistrationClusterID, "raft-registration-cluster-id", "", 104, "The raft cluster id assigned to the registration database")

	// Datastore
	clusterCmd.PersistentFlags().StringVarP(&DataStoreEngine, "datastore", "", "raft", "Data store type [ memory | sqlite | mysql | postgres | cockroach | raft ]")
	clusterCmd.PersistentFlags().StringVarP(&DeviceDataStore, "data-store", "", "raft", "Where to store historical device data [ raft | gorm | redis ]")

	// Default cluster parameters (for provisioning new accounts)
	clusterCmd.PersistentFlags().IntVarP(&DefaultConsistencyLevel, "default-consistency-level", "", common.CONSISTENCY_LOCAL, "CONSISTENCY_LOCAL or CONSISTENCY_QUORUM")
	clusterCmd.PersistentFlags().IntVarP(&DefaultConfigStoreType, "default-config-store", "", config.RAFT_DISK_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	clusterCmd.PersistentFlags().IntVarP(&DefaultStateStoreType, "default-state-store", "", config.RAFT_MEMORY_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	clusterCmd.PersistentFlags().IntVarP(&DefaultDataStoreType, "default-data-store", "", config.GORM_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	//clusterCmd.PersistentFlags().IntVarP(&DefaultDataStoreType, "default-data-store", "", config.RAFT_DISK_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")

	viper.BindPFlags(clusterCmd.PersistentFlags())

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

		App.Mode = Mode
		App.Mailer = service.NewMailer(App, nil)
		App.IdGenerator = util.NewIdGenerator(DataStoreEngine)
		App.IdSetter = util.NewIdSetter(App.IdGenerator)
		App.DefaultConsistencyLevel = viper.GetInt("default-consistency-level")
		App.DefaultConfigStoreType = viper.GetInt("default-config-store")
		App.DefaultStateStoreType = viper.GetInt("default-state-store")
		App.DefaultDataStoreType = viper.GetInt("default-data-store")
		App.ClusterID = ClusterID

		if ClusterRaft != "" {
			pieces := strings.Split(ClusterRaft, ",")
			ClusterBootstrap = len(pieces)
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

		raftOptions := clusterutil.RaftOptions{
			Port:                  ClusterRaftPort,
			RequestedLeaderID:     ClusterRaftLeaderID,
			SystemClusterID:       ClusterID,
			OrganizationClusterID: RaftOrganizationClusterID,
			UserClusterID:         RaftUserClusterID,
			RoleClusterID:         RaftRoleClusterID,
			AlgorithmClusterID:    RaftAlgorithmClusterID,
			RegistrationClusterID: RaftRegistrationClusterID}

		farmProvisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
		farmDeprovisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
		farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)

		params := clusterutil.NewClusterParams(App.Logger, raftOptions, NodeID,
			ClusterIaasProvider, ClusterRegion, ClusterZone, App.DataDir, localAddress,
			ClusterListenAddress, gossipPeers, raftPeers, ClusterJoin, ClusterGossipPort,
			ClusterRaftPort, ClusterRaftLeaderID, ClusterVirtualNodes, ClusterMaxNodes,
			ClusterBootstrap, DatabaseInit, App.IdGenerator, App.IdSetter, farmProvisionerChan,
			farmDeprovisionerChan, farmTickerProvisionerChan)

		gossipNode := cluster.NewGossipNode(params, clusterutil.NewHashring(ClusterVirtualNodes))
		gossipNode.Join()
		go gossipNode.Run()

		raftNode := gossipNode.GetSystemRaft()
		for raftNode == nil {
			App.Logger.Info("Waiting for enough nodes to build the Raft quorum...")
			time.Sleep(1 * time.Second)
			raftNode = gossipNode.GetSystemRaft()
		}

		//App.ServerStore = cluster.NewRaftFarmConfigStore(App.Logger, App.RaftCluster)
		//App.FarmStore = cluster.NewRaftFarmStateStore(App.Logger, App.RaftCluster)

		// TODO: Check device hardware and firmware versions at startup, update devices db table
		builder := builder.NewClusterConfigBuilder(App, params,
			gossipNode, raftNode, DeviceDataStore, AppStateTTL, AppStateTick)
		rsaKeyPair, serviceRegistry, restServices, _, err := builder.Build()
		if err != nil {
			App.Logger.Fatal(err)
		}

		App.KeyPair = rsaKeyPair
		//App.Server = serverConfig.(*config.Server)
		//App.DeviceIndex = deviceIndex
		//App.ChannelIndex = channelIndex

		for _, farmService := range serviceRegistry.GetFarmServices() {
			go farmService.RunCluster()
		}

		if changefeedService := serviceRegistry.GetChangefeedService(); changefeedService != nil {
			changefeedService.Subscribe()
		}

		webserver := webservice.NewWebserver(App, gossipNode, raftNode,
			serviceRegistry, restServices, farmTickerProvisionerChan)
		go webserver.Run()
		go webserver.RunClusterProvisionerConsumer()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT) // syscall.SIGTERM, syscall.SIGHUP)

		<-stop

		App.Logger.Info("Shutting down...")

		//webserver.Shutdown()

		if err := gossipNode.Shutdown(); err != nil {
			App.Logger.Error(err)
		}

		if err := raftNode.Shutdown(); err != nil {
			App.Logger.Fatal(err)
		}

		App.Logger.Info("Shutdown complete")
	},
}
