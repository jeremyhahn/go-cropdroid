//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/util"
)

const (
	TestSuiteName         = "datastore_raft"
	TestDataDir           = "test-data"
	OrganizationClusterID = uint64(100)
	FarmConfigClusterID   = uint64(101)
	FarmStateClusterID    = uint64(102)
	DeviceConfigClusterID = uint64(103)
	UserClusterID         = uint64(104)
	RoleClusterID         = uint64(105)
	AlgorithmClusterID    = uint64(106)
	RegistrationClusterID = uint64(107)
	DeviceStateClusterID  = uint64(108)
	DeviceDataClusterID   = uint64(109)
	NodeCount             = 3
	RaftLeaderID          = 3
	NodeID                = 1
)

var (
	ConcurrentTestCounter                = 0
	ClusterID                            = uint64(420) // system / gossip raft
	IntegrationTestCluster *LocalCluster = nil
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	IntegrationTestCluster = NewClusterIntegrationTest()
	IntegrationTestCluster.DestroyData()
	IntegrationTestCluster.StartCluster()
}

func teardown() {
}

func NewClusterIntegrationTest() *LocalCluster {

	if NodeCount < 3 {
		panic("NodeCount must be greater than 3")
	}

	if NodeCount > 9 {
		panic("NodeCount must be less than 9")
	}

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}

	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	app := &app.App{
		Logger:      logger,
		Location:    location,
		NodeID:      NodeID,
		ClusterID:   ClusterID,
		DataDir:     fmt.Sprintf("./%s", TestDataDir),
		IdGenerator: idGenerator,
		IdSetter:    util.NewIdSetter(idGenerator)}

	return NewLocalCluster(app, 3, ClusterID)
}
