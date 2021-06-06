// +build broken

package test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

var CurrentTest *DatastoreTest = &DatastoreTest{mutex: &sync.Mutex{}}

type DatastoreTest struct {
	mutex    *sync.Mutex
	db       gormstore.GormDB
	gorm     *gorm.DB
	logger   *logging.Logger
	location *time.Location
}

func NewIntegrationTest() *DatastoreTest {

	CurrentTest.mutex.Lock()

	appName := "cropdroid-datstore-test"

	backend, _ := logging.NewSyslogBackend(appName)
	logging.SetBackend(backend)
	logger := logging.MustGetLogger(appName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}

	database := gormstore.NewGormDB(logger, &gormstore.GormInitParams{
		DataDir:  "/tmp",
		DBName:   appName,
		Location: location})
	gormdb := database.Connect()
	gormdb.LogMode(true)

	//database.Migrate()

	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = location
	return CurrentTest
}

func (dt *DatastoreTest) Cleanup() {
	if CurrentTest != nil {
		CurrentTest.db.Close()
		CurrentTest.db.Drop()
		CurrentTest.mutex.Unlock()
	}
}

func NewUnitTestContext() (*app.App, common.Scope) {
	wd, _ := os.Getwd()

	logger := logging.MustGetLogger(common.APPNAME)
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logging.SetLevel(logging.DEBUG, "")

	location, _ := time.LoadLocation("America/New_York")

	_app := &app.App{
		HomeDir:  wd,
		Logger:   logger,
		Location: location}

	farmID := 0
	controllerType := "testcontroller"
	serverConfig, controllerState := createTestFarm(farmID, controllerType)

	farmStateMap := make(map[string]common.ControllerStateMap)
	farmStateMap[controllerType] = controllerState
	farmState := common.CreateFarmState(farmStateMap)

	scope := common.CreateScope(logger, serverConfig, farmState, &model.User{
		ID:       1,
		Email:    "root@localhost",
		Password: "$ecret"},
		location)

	return _app, scope
}

func createTestFarm(farmID int, controllerType string) (config.FarmConfig, common.ControllerStateMap) {

	testChannelID := 0
	fakeValue := 75.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	controllerState := common.CreateControllerStateMap(metricMap, []int{0})

	metrics := []config.MetricConfig{
		&config.Metric{
			ID:  1,
			Key: "humidity0"}}

	condition := &config.Condition{
		ID:         1,
		ChannelID:  1,
		MetricID:   1,
		Comparator: ">",
		Threshold:  55.0}

	channels := []config.ChannelConfig{
		&config.Channel{
			ID:           1,
			ControllerID: 1,
			ChannelID:    testChannelID,
			Name:         "test",
			Enable:       true,
			Notify:       true,
			Conditions:   []config.ConditionConfig{condition}}}

	controller := &config.Controller{
		ID:   1,
		Type: controllerType,
		Configs: map[string]string{
			fmt.Sprintf("%s.notify", controllerType): "true",
		},
		Metrics:  metrics,
		Channels: channels}

	farmConfig := &config.Farm{
		ID:          farmID,
		OrgID:       0,
		Mode:        "virtual",
		Name:        "Test Farm",
		Interval:    50,
		Controllers: []config.ControllerConfig{controller}}
	/*
		serverConfig := &config.Server{
			ID:       1,
			OrgID:    0,
			Interval: 60,
			Mode:     "server",
			Smtp:     nil,
			Farms:    []config.FarmConfig{farmConfig}}

		return serverConfig, controllerState
	*/
	return farmConfig, controllerState
}

/*
func getConfigFile(logger *logging.Logger) *common.Config {
	configYaml := fmt.Sprintf("%s/config.yaml", "../test")
	configData, err := common.NewConfigFile(configYaml).Read()
	if err != nil {
		logger.Fatalf("Unable to load configuration file %s: %s", configYaml, err)
	}
	configFile := &common.Config{}
	err = yaml.Unmarshal(configData, &configFile)
	if err != nil {
		logger.Fatalf("Unable to unmarshal config: %s", err)
	}
	logger.Debugf("Loaded config: %s", configYaml)
	return configFile
}
*/
