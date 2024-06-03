//go:build cluster
// +build cluster

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
	RaftCustomerClusterID     uint64
)

func init() {

	// IaaS
	clusterCmd.PersistentFlags().StringVarP(&ClusterIaasProvider, "provider", "", "kvm", "Infrastructure-as-a-Service (IaaS) provider [ libvirt | terraform ]")

	// General cluster
	clusterCmd.PersistentFlags().Uint64VarP(&App.ClusterID, "cluster-id", "", 1, "Cluster unique identifier")
	clusterCmd.PersistentFlags().IntVarP(&ClusterVirtualNodes, "vnodes", "", 64, "Number of virtual nodes on the consistent hash ring")
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
	clusterCmd.PersistentFlags().Uint64VarP(&RaftCustomerClusterID, "raft-customer-cluster-id", "", 105, "The raft cluster id assigned to the customer database")

	// Datastore
	clusterCmd.PersistentFlags().StringVarP(&DataStoreEngine, "datastore", "", "raft", "Data store type [ memory | sqlite | mysql | postgres | cockroach | raft ]")
	clusterCmd.PersistentFlags().StringVarP(&DeviceDataStore, "data-store", "", "raft", "Where to store historical device data [ raft | gorm | redis ]")

	// Default cluster parameters (for provisioning new accounts)
	clusterCmd.PersistentFlags().IntVarP(&App.DefaultConsistencyLevel, "default-consistency-level", "", common.CONSISTENCY_LOCAL, "CONSISTENCY_LOCAL or CONSISTENCY_QUORUM")
	clusterCmd.PersistentFlags().IntVarP(&App.DefaultConfigStoreType, "default-config-store", "", config.RAFT_DISK_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	clusterCmd.PersistentFlags().IntVarP(&App.DefaultStateStoreType, "default-state-store", "", config.RAFT_MEMORY_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	clusterCmd.PersistentFlags().IntVarP(&App.DefaultDataStoreType, "default-data-store", "", config.GORM_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")
	//clusterCmd.PersistentFlags().IntVarP(&DefaultDataStoreType, "default-data-store", "", config.RAFT_DISK_STORE, "RAFT_DISK_STORE, RAFT_MEMORY_STORE, GORM_STORE")

	viper.BindPFlags(clusterCmd.PersistentFlags())

	rootCmd.AddCommand(clusterCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Run the cropdroid service in cluster mode",
	Long: `Starts the cropdroid real-time protection and notification service
	using a cluster license. The Raft consenus algorithm is combined with an
	embedded rocksdb database to provide a high performance, highly available,
	fault-tolerant, and massively scalable cloud native architecture.`,
	Run: func(cmd *cobra.Command, args []string) {

		sigChan := make(chan os.Signal, 1)

		App.IdGenerator = util.NewIdGenerator(DataStoreEngine)
		App.IdSetter = util.NewIdSetter(App.IdGenerator)

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
			RegistrationClusterID: RaftRegistrationClusterID,
			CustomerClusterID:     RaftCustomerClusterID}

		farmProvisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
		farmDeprovisionerChan := make(chan config.Farm, common.BUFFERED_CHANNEL_SIZE)
		farmTickerProvisionerChan := make(chan uint64, common.BUFFERED_CHANNEL_SIZE)

		params := clusterutil.NewClusterParams(
			App.Logger,
			raftOptions,
			App.NodeID,
			ClusterIaasProvider,
			ClusterRegion,
			ClusterZone,
			App.DataDir,
			localAddress,
			ClusterListenAddress,
			gossipPeers,
			raftPeers,
			ClusterJoin,
			ClusterGossipPort,
			ClusterRaftPort,
			ClusterRaftLeaderID,
			ClusterVirtualNodes,
			ClusterMaxNodes,
			ClusterBootstrap,
			DatabaseInit,
			App.IdGenerator,
			App.IdSetter,
			farmProvisionerChan,
			farmDeprovisionerChan,
			farmTickerProvisionerChan)

		gossipNode := cluster.NewGossipNode(params, clusterutil.NewHashring(ClusterVirtualNodes))
		gossipNode.Join()
		go gossipNode.Run()

		raftNode := gossipNode.GetSystemRaft()
		for raftNode == nil {
			App.Logger.Info("Waiting for enough nodes to build the Raft quorum...")
			time.Sleep(1 * time.Second)
			raftNode = gossipNode.GetSystemRaft()
		}
		App.ClusterID = raftNode.GetParams().ClusterID
		App.NodeID = raftNode.GetParams().NodeID

		// TODO: Check device hardware and firmware versions at startup, update devices db table
		builder := builder.NewClusterConfigBuilder(
			App,
			params,
			gossipNode,
			raftNode,
			DeviceDataStore,
			App.StateTTL,
			App.StateTick)
		mapperRegistry, serviceRegistry, restServiceRegistry, _, err := builder.Build()
		if err != nil {
			App.Logger.Fatal(err)
		}

		//App.Server = serverConfig.(*config.Server)
		//App.DeviceIndex = deviceIndex
		//App.ChannelIndex = channelIndex

		for _, farmService := range serviceRegistry.GetFarmServices() {
			go farmService.RunCluster()
		}

		webserver := webservice.NewClusterWebServerV1(
			App,
			gossipNode,
			raftNode,
			mapperRegistry,
			serviceRegistry,
			restServiceRegistry,
			farmTickerProvisionerChan)
		go webserver.Run()
		go webserver.RunClusterProvisionerConsumer()

		serviceRegistry.GetEventLogService(ClusterID).Create(ClusterID, common.CONTROLLER_TYPE_SERVER, "System", "Startup")

		signal.Notify(sigChan, syscall.SIGINT) // catch CTRL+C // syscall.SIGTERM, syscall.SIGHUP)

		<-App.ShutdownChan
		close(App.ShutdownChan)
		close(sigChan)

		serviceRegistry.GetEventLogService(ClusterID).Create(ClusterID, common.CONTROLLER_TYPE_SERVER, "System", "Shutdown")

		webserver.Shutdown()

		if err := gossipNode.Shutdown(); err != nil {
			App.Logger.Error(err)
		}

		if err := raftNode.Shutdown(); err != nil {
			App.Logger.Fatal(err)
		}

		App.Logger.Info("Graceful shutdown complete")
	},
}
