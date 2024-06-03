package gorm

import (
	"fmt"
	"os"
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/util"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//_ "gorm.io/driver/mysql"
	//_ "gorm.io/driver/postgres"
	// "gorm.io/driver/sqlserver"
	// "gorm.io/driver/clickhouse"

	logging "github.com/op/go-logging"
)

type GormDB interface {
	GORM() *gorm.DB
	Create() error
	Connect(serverConnection bool) *gorm.DB
	CloneConnection() *gorm.DB
	Migrate() error
	Drop()
}

type GormInitParams struct {
	AppMode           string
	DebugFlag         bool
	EnableDefaultFarm bool
	DataDir           string
	Engine            string
	Path              string
	Host              string
	Port              int
	Username          string
	Password          string
	CACert            string
	TLSKey            string
	TLSCert           string

	DBName   string
	Location *time.Location
}

type GormDatabase struct {
	logger             *logging.Logger
	params             *GormInitParams
	db                 *gorm.DB
	isServerConnection bool
	isInMemoryDatabase bool
	GormDB
}

func NewGormDB(logger *logging.Logger, params *GormInitParams) GormDB {
	return &GormDatabase{logger: logger, params: params}
}

// Connect creates and returns a new connection to the database
func (database *GormDatabase) Connect(serverConnection bool) *gorm.DB {
	//database.logger.Debug(fmt.Sprintf("Datastore: %s", database.params.Engine))
	database.isServerConnection = serverConnection
	switch database.params.Engine {
	case "sqlite-memory":
		database.isInMemoryDatabase = true
		database.db = database.newSQLite(fmt.Sprintf("file:%s?mode=memory&cache=shared", database.params.DBName))
	case "sqlite":
		sqlite := fmt.Sprintf("%s/%s.db", database.params.DataDir, database.params.DBName)
		exists, err := util.FileExists(sqlite)
		if err != nil {
			database.logger.Error(err)
			return nil
		}
		if !exists {
			os.MkdirAll(database.params.DataDir, 0755)
			os.Create(sqlite)
		}
		database.db = database.newSQLite(sqlite)
	// case "cockroach":
	// 	database.db = database.newCockroachDB()
	// case "postgres":
	// 	database.db = database.newPgsql()
	// case "mysql":
	// 	database.db = database.newMySQL(serverConnection)
	default:
		database.logger.Fatalf("[gorm.Connect] Unsupported GORM engine: %s", database.params.Engine)
	}

	return database.db
}

// Uses the connection parameters and credentials from the current
// database session to establish a new connection.
func (database *GormDatabase) CloneConnection() *gorm.DB {
	// if database.isInMemoryDatabase {
	// 	return database.db
	// }
	return database.Connect(database.isServerConnection)
}

// Returns the underlying GORM handle
func (database *GormDatabase) GORM() *gorm.DB {
	return database.db
}

// Create a new database
func (database *GormDatabase) Create() error {
	if database.params.Engine != "sqlite" && database.params.Engine != "memory" {
		query := fmt.Sprintf("CREATE DATABASE %s;", database.params.DBName)
		return database.db.Exec(query).Error
	}
	return nil
}

// Migrate will import / alter the current schema to match entities defined in config package
func (database *GormDatabase) Migrate() error {

	database.db.AutoMigrate(config.AlgorithmStruct{})
	database.db.AutoMigrate(config.ChannelStruct{})
	database.db.AutoMigrate(config.ConditionStruct{})
	database.db.AutoMigrate(config.DeviceSettingStruct{})
	database.db.AutoMigrate(config.DeviceStruct{})
	database.db.AutoMigrate(config.FarmStruct{})
	database.db.AutoMigrate(config.ServerLicenseStruct{})
	database.db.AutoMigrate(config.OrganizationLicenseStruct{})
	database.db.AutoMigrate(config.FarmLicenseStruct{})
	database.db.AutoMigrate(config.MetricStruct{})
	database.db.AutoMigrate(config.OrganizationStruct{})
	database.db.AutoMigrate(config.PermissionStruct{})
	database.db.AutoMigrate(config.RegistrationStruct{})
	database.db.AutoMigrate(config.RoleStruct{})
	database.db.AutoMigrate(config.ScheduleStruct{})
	//database.db.AutoMigrate(&config.Server{})
	database.db.AutoMigrate(config.UserStruct{})
	database.db.AutoMigrate(config.WorkflowStepStruct{})
	database.db.AutoMigrate(config.WorkflowStruct{})
	// Entities
	database.db.AutoMigrate(entity.EventLog{})
	database.db.AutoMigrate(entity.InventoryType{})
	database.db.AutoMigrate(entity.Inventory{})

	// Billing & Shopping Cart
	database.db.AutoMigrate(config.CustomerStruct{})
	database.db.AutoMigrate(config.AddressStruct{})
	database.db.AutoMigrate(config.ShippingAddressStruct{})

	return nil
}

// Drop the database
func (database *GormDatabase) Drop() {
	switch database.params.Engine {
	case "sqlite":
		os.Remove(fmt.Sprintf("%s/%s.db", database.params.DataDir, database.params.DBName))
	case "postgres", "cockroach":
		query := fmt.Sprintf("DROP DATABASE %s CASCADE", database.params.DBName)
		database.logger.Debug(query)
		database.db.Exec(query)
	case "mysql":
		query := fmt.Sprintf("DROP DATABASE %s;", database.params.DBName)
		database.logger.Debug(query)
		database.db.Exec(query)
	}
}

// Create a new sqlite database connection
func (database *GormDatabase) newSQLite(dbname string) *gorm.DB {
	gormConfig := &gorm.Config{}
	if database.params.DebugFlag {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}
	db, err := gorm.Open(sqlite.Open(dbname), gormConfig)
	if err != nil {
		database.logger.Fatalf("SQLite Error: %s", err)
	}
	// db.Exec("PRAGMA foreign_keys = OFF")
	// db.Exec("PRAGMA journal_mode = WAL")
	// db.Exec("PRAGMA ignore_check_constraints = ON")
	return db
}

// // Create a new mysql database connection
// func (database *GormDatabase) newMySQL(serverConnection bool) *gorm.DB {
// 	var connStr string
// 	if serverConnection {
// 		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&parseTime=True&loc=Local",
// 			database.params.Username, database.params.Password, database.params.Host, database.params.Port)
// 	} else {
// 		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
// 			database.params.Username, database.params.Password, database.params.Host, database.params.Port, database.params.DBName)
// 	}
// 	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		database.logger.Fatalf("MySQL Error: %s", err)
// 	}
// 	return db
// }

// // Create a new postgres database connection
// func (database *GormDatabase) newPgsql() *gorm.DB {
// 	// sslmode=disable TimeZone=America/New_York
// 	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
// 		database.params.Host, database.params.Port, database.params.Username, database.params.DBName, database.params.Password)
// 	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		database.logger.Fatalf("Postgres Error: %s", err)
// 	}
// 	return db
// }

// // Create a new cockroach db connection
// func (database *GormDatabase) newCockroachDB() *gorm.DB {
// 	sslParams := "sslmode=disable"
// 	if database.params.CACert != "" && database.params.TLSKey != "" && database.params.TLSCert != "" {
// 		sslParams = fmt.Sprintf("sslmode=require&sslkey=%s&sslcert=%s&sslrootcert=%s",
// 			database.params.TLSKey, database.params.TLSCert, database.params.CACert)
// 	}
// 	connStr := fmt.Sprintf("postgres://%s@%s:%d/%s?%s",
// 		database.params.Username, database.params.Host,
// 		database.params.Port, database.params.DBName, sslParams)
// 	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
// 	if err != nil {
// 		database.logger.Fatalf("CockroachDB Error: %s", err)
// 	}
// 	return db
// }
