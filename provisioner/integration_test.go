package provisioner

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

var CurrentTest *ProvisionerTest = &ProvisionerTest{mutex: &sync.Mutex{}}
var Location *time.Location
var TestSuiteName = "cropdroid_service_test"

type ProvisionerTest struct {
	mutex    *sync.Mutex
	app      *app.App
	db       gormstore.GormDB
	gorm     *gorm.DB
	logger   *logging.Logger
	location *time.Location
}

func NewIntegrationTest() *ProvisionerTest {

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

	gormdb := database.Connect(true)
	database.Create()
	gormdb.Close()

	gormdb = database.Connect(false)
	database.Migrate()

	app := &app.App{
		GORM:           gormdb,
		GORMInitParams: gormInitParams,
		KeyDir:         "../keys",
		Logger:         logger,
		Location:       Location}

	CurrentTest.app = app
	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = Location
	return CurrentTest
}

func (dt *ProvisionerTest) Cleanup() {
	if CurrentTest != nil {
		// Close app user connection
		CurrentTest.db.Close()

		// Connect as server admin, drop db
		CurrentTest.db.Connect(true)
		CurrentTest.db.Drop()
		CurrentTest.db.Close()

		CurrentTest.mutex.Unlock()
	}
}

func createSqliteParams() *gormstore.GormInitParams {
	return &gormstore.GormInitParams{
		DebugFlag: true,
		DataDir:   "/tmp",
		Engine:    "sqlite",
		DBName:    TestSuiteName,
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
