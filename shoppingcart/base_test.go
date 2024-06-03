package shoppingcart

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var CurrentTest *TestSuite = &TestSuite{mutex: &sync.Mutex{}}
var Location *time.Location
var TestSuiteName = "cropdroid_shoppingcart_test"
var SKIP_TEARDOWN_FLAG = false

// This user id maps to root@test.com
// Stripe does not recommend submitting card details from the server to their API
// beacuse they can't control what the server does with the card number prior
// to sending it to stripe. That make testing hard since the tests that require
// a card need the customer in stripe, with a successful payment already made, and
// chose to save their card for future use.
var TEST_CUSTOMER_ID = uint64(2060288868)

type TestSuite struct {
	mutex    *sync.Mutex
	app      *app.App
	db       gormstore.GormDB
	gorm     *gorm.DB
	logger   *logging.Logger
	location *time.Location
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {

	CurrentTest.mutex.Lock()

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	Location = location

	gormInitParams := createSqliteParams() //createMemoryParams(), createCockroachParams()
	database := gormstore.NewGormDB(logger, gormInitParams)

	database.Connect(true)
	database.Create()

	gormdb := database.Connect(false)
	database.Migrate()

	app := &app.App{
		GORMInitParams: gormInitParams,
		CertDir:        "../db/certs",
		Logger:         logger,
		Location:       Location}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("/etc/cropdroid/"))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.cropdroid/"))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		app.Logger.Errorf("%s", err)
	}

	viper.Unmarshal(app)

	CurrentTest.app = app
	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = Location
}

func teardown() {
	if SKIP_TEARDOWN_FLAG {
		CurrentTest.logger.Debug("skipping database teardown and resetting flag for next test")
		SKIP_TEARDOWN_FLAG = false
		return
	}
	if CurrentTest != nil {
		// Connect as server admin, drop db
		CurrentTest.db.Connect(true)
		CurrentTest.db.Drop()

		CurrentTest.mutex.Unlock()
	}
}

func createSqliteParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		DataDir:   "../db",
		Engine:    "sqlite",
		DBName:    "cropdroid",
		Location:  Location}
}

func createMemoryParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		DataDir:   "/tmp",
		Engine:    "memory",
		DBName:    TestSuiteName,
		Location:  Location}
}

func createCockroachParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		Engine:    "cockroach",
		Host:      "localhost",
		Port:      26257,
		Username:  "root",
		//Password:  "dev",
		DBName:   TestSuiteName,
		Location: Location}
}

func createMyqlParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		Engine:    "mysql",
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "dev",
		DBName:    TestSuiteName,
		Location:  Location}
}

func createPostgresParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		Engine:    "postgres",
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "dev",
		DBName:    TestSuiteName,
		Location:  Location}
}
