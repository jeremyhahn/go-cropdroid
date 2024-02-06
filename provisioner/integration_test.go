package provisioner

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

var CurrentTest *ProvisionerTest = &ProvisionerTest{mutex: &sync.Mutex{}}
var Location *time.Location
var TestSuiteName = "cropdroid_provisioner_test"

type ProvisionerTest struct {
	mutex          *sync.Mutex
	db             gormstore.GormDB
	gorm           *gorm.DB
	logger         *logging.Logger
	location       *time.Location
	idGenerator    util.IdGenerator
	passwordHasher util.PasswordHasher
	mode           string
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

	gormdb = database.Connect(false)
	database.Migrate()

	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = Location
	CurrentTest.passwordHasher = util.NewPasswordHasher()
	CurrentTest.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	return CurrentTest
}

func (dt *ProvisionerTest) Cleanup() {
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
