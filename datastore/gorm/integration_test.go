package gorm

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

var CurrentTest *DatastoreTest = &DatastoreTest{mutex: &sync.Mutex{}}
var Location *time.Location
var TestSuiteName = "cropdroid_datastore_test"

type DatastoreTest struct {
	mutex       *sync.Mutex
	db          GormDB
	gorm        *gorm.DB
	logger      *logging.Logger
	location    *time.Location
	idGenerator util.IdGenerator
}

func NewIntegrationTest() *DatastoreTest {

	CurrentTest.mutex.Lock()

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	Location = location

	os.Remove(fmt.Sprintf("/tmp/%s.db", TestSuiteName))

	//database := NewGormDB(logger, createMemoryParams())
	database := NewGormDB(logger, createSqliteParams())
	//database := NewGormDB(logger, createCockroachParams())

	gormdb := database.Connect(true)
	database.Create()
	gormdb.Close()

	gormdb = database.Connect(false)
	//database.Migrate()

	CurrentTest.db = database
	CurrentTest.gorm = gormdb
	CurrentTest.logger = logger
	CurrentTest.location = Location
	CurrentTest.idGenerator = util.NewIdGenerator(common.DATASTORE_TYPE_32BIT)
	return CurrentTest
}

func (dt *DatastoreTest) Cleanup() {
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

func createSqliteParams() *GormInitParams {
	return &GormInitParams{
		DebugFlag: true,
		DataDir:   "/tmp",
		Engine:    "sqlite",
		DBName:    TestSuiteName,
		Location:  Location}
}

func createMemoryParams() *GormInitParams {
	return &GormInitParams{
		DebugFlag: true,
		DataDir:   "/tmp",
		Engine:    "memory",
		DBName:    TestSuiteName,
		Location:  Location}
}

func createCockroachParams() *GormInitParams {
	return &GormInitParams{
		DebugFlag: true,
		Engine:    "cockroach",
		Host:      "localhost",
		Port:      26257,
		Username:  "root",
		//Password:  "dev",
		DBName:   TestSuiteName,
		Location: Location}
}

func createMyqlParams() *GormInitParams {
	return &GormInitParams{
		DebugFlag: true,
		Engine:    "mysql",
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "dev",
		DBName:    TestSuiteName,
		Location:  Location}
}

func createPostgresParams() *GormInitParams {
	return &GormInitParams{
		DebugFlag: true,
		Engine:    "postgres",
		Host:      "localhost",
		Port:      3306,
		Username:  "root",
		Password:  "dev",
		DBName:    TestSuiteName,
		Location:  Location}
}
