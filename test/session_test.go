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
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

var CurrentTest *ServiceTest = &ServiceTest{mutex: &sync.Mutex{}}

type ServiceTest struct {
	mutex    *sync.Mutex
	db       gormstore.GormDB
	gorm     *gorm.DB
	logger   *logging.Logger
	location *time.Location
}

func NewIntegrationTest() *ServiceTest {

	CurrentTest.mutex.Lock()

	appName := "cropdroid-service-test"

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
	gormdb := database.Connect(false)
	gormdb.LogMode(true)

	//database.Migrate()

	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = location
	return CurrentTest
}

func (dt *ServiceTest) Cleanup() {
	if CurrentTest != nil {
		CurrentTest.db.Close()
		CurrentTest.db.Drop()
		CurrentTest.mutex.Unlock()
	}
}

func NewUnitTestSession() (*app.App, service.Session) {
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

	// farmID := 0
	// deviceType := "testdevice"
	// serverConfig, deviceState := createTestFarm(farmID, deviceType)

	// farmStateMap := make(map[string]state.FarmStateMap)
	// farmStateMap[deviceType] = deviceState
	// farmState := state.CreateFarmState(farmID, deviceState)

	//session := service.CreateSession(logger, farmService, &model.User{
	session := service.CreateSession(logger, nil, nil, &model.User{
		ID:       1,
		Email:    "root@localhost",
		Password: "$ecret"})

	return _app, session
}

func createTestFarm(farmID uint64, deviceType string) (config.FarmConfig, state.DeviceStateMap) {

	testChannelID := 0
	fakeValue := 75.0

	metricMap := make(map[string]float64, 0)
	metricMap["humidity0"] = fakeValue
	deviceState := state.CreateDeviceStateMap(metricMap, []int{0})

	metrics := []config.Metric{
		config.Metric{
			ID:  1,
			Key: "humidity0"}}

	condition := &config.Condition{
		ID:         1,
		ChannelID:  1,
		MetricID:   1,
		Comparator: ">",
		Threshold:  55.0}

	channels := []config.Channel{
		config.Channel{
			ID:         1,
			DeviceID:   1,
			ChannelID:  testChannelID,
			Name:       "test",
			Enable:     true,
			Notify:     true,
			Conditions: []config.Condition{*condition}}}

	device := config.Device{
		ID:   1,
		Type: deviceType,
		// Configs: map[string]string{
		// 	fmt.Sprintf("%s.notify", deviceType): "true",
		// },
		Configs: []config.DeviceConfigItem{
			config.DeviceConfigItem{
				Key:   fmt.Sprintf("%s.notify", deviceType),
				Value: "true"},
		},
		Metrics:  metrics,
		Channels: channels}

	farmConfig := &config.Farm{
		ID:             farmID,
		OrganizationID: 0,
		Mode:           "virtual",
		Name:           "Test Farm",
		Interval:       50,
		Devices:        []config.Device{device}}

	/*
		serverConfig := &config.Server{
			ID:       1,
			OrgID:    0,
			Interval: 60,
			Mode:     "server",
			Smtp:     nil,
			Farms:    []config.FarmConfig{farmConfig}}

		return serverConfig, deviceState
	*/
	return farmConfig, deviceState
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
