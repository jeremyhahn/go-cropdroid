// +build cluster

package gorm

import (
	"fmt"
	"os"
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/mattn/go-sqlite3"
	logging "github.com/op/go-logging"
)

type GormDB interface {
	Create() error
	Connect(serverConnection bool) *gorm.DB
	Migrate() error
	Drop()
	Close()
}

type GormInitParams struct {
	DebugFlag bool
	DataDir   string
	Engine    string
	Path      string
	Host      string
	Port      int
	Username  string
	Password  string
	CACert    string
	TLSKey    string
	TLSCert   string

	DBName   string
	Location *time.Location
}

type GormDatabase struct {
	logger *logging.Logger
	params *GormInitParams
	db     *gorm.DB
	GormDB
}

func NewGormDB(logger *logging.Logger, params *GormInitParams) GormDB {
	return &GormDatabase{logger: logger, params: params}
}

// Connect creates and returns a new connection to the database
func (database *GormDatabase) Connect(serverConnection bool) *gorm.DB {
	database.logger.Debug(fmt.Sprintf("Connecting to datastore: %s", database.params.Engine))
	switch database.params.Engine {
	case "memory":
		//"file:%s?mode=memory&cache=shared"
		database.db = database.newSQLite(fmt.Sprintf("file:%s?mode=memory", database.params.DBName))
		//database.db.LogMode(true)
		if err := NewGormClusterInitializer(database.logger, database.db, database.params.Location).Initialize(); err != nil {
			database.logger.Fatal(err)
		}
	case "sqlite":
		sqlite := fmt.Sprintf("%s/%s.db", database.params.DataDir, database.params.DBName)
		database.db = database.newSQLite(sqlite)
	case "cockroach":
		database.db = database.newCockroachDB()
	case "postgres":
		database.db = database.newPgsql()
	case "mysql":
		database.db = database.newMySQL(serverConnection)
	default:
		database.logger.Fatalf("[gorm.Connect] Unsupported GORM engine: %s", database.params.Engine)
	}
	database.db.LogMode(database.params.DebugFlag)
	return database.db
}

// Create a new database
func (database *GormDatabase) Create() error {
	if database.params.Engine != "sqlite" {
		query := fmt.Sprintf("CREATE DATABASE %s;", database.params.DBName)
		return database.db.Exec(query).Error
	}
	return nil
}

// Migrate will import / alter the current schema to match entities defined in config package
func (database *GormDatabase) Migrate() error {
	database.db.AutoMigrate(&config.Permission{})
	database.db.AutoMigrate(&config.User{})
	database.db.AutoMigrate(&config.Role{})
	database.db.AutoMigrate(&config.Device{})
	database.db.AutoMigrate(&config.DeviceConfigItem{})
	database.db.AutoMigrate(&config.Metric{})
	database.db.AutoMigrate(&config.Channel{})
	database.db.AutoMigrate(&config.Condition{})
	database.db.AutoMigrate(&config.Algorithm{})
	database.db.AutoMigrate(&entity.EventLog{})
	database.db.AutoMigrate(&config.Schedule{})
	database.db.AutoMigrate(&config.Farm{})
	database.db.AutoMigrate(&config.Organization{})
	database.db.AutoMigrate(&config.License{})
	database.db.AutoMigrate(&entity.InventoryType{})
	database.db.AutoMigrate(&entity.Inventory{})
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

// Close the database connection
func (database *GormDatabase) Close() {
	database.db.Close()
}

// Create a new sqlite database connection
func (database *GormDatabase) newSQLite(dbname string) *gorm.DB {
	db, err := gorm.Open("sqlite3", dbname)
	if err != nil {
		database.logger.Fatalf("SQLite Error: %s", err)
	}
	return db
}

// Create a new mysql database connection
func (database *GormDatabase) newMySQL(serverConnection bool) *gorm.DB {
	var connStr string
	if serverConnection {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&parseTime=True&loc=Local",
			database.params.Username, database.params.Password, database.params.Host, database.params.Port)
	} else {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			database.params.Username, database.params.Password, database.params.Host, database.params.Port, database.params.DBName)
	}
	db, err := gorm.Open("mysql", connStr)
	if err != nil {
		database.logger.Fatalf("MySQL Error: %s", err)
	}
	return db
}

// Create a new postgres database connection
func (database *GormDatabase) newPgsql() *gorm.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		database.params.Host, database.params.Port, database.params.Username, database.params.DBName, database.params.Password)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		database.logger.Fatalf("Postgres Error: %s", err)
	}
	return db
}

// Create a new cockroach db connection
func (database *GormDatabase) newCockroachDB() *gorm.DB {
	sslParams := "sslmode=disable"
	if database.params.CACert != "" && database.params.TLSKey != "" && database.params.TLSCert != "" {
		sslParams = fmt.Sprintf("sslmode=require&sslkey=%s&sslcert=%s&sslrootcert=%s",
			database.params.TLSKey, database.params.TLSCert, database.params.CACert)
	}
	connStr := fmt.Sprintf("postgres://%s@%s:%d/%s?%s",
		database.params.Username, database.params.Host,
		database.params.Port, database.params.DBName, sslParams)

	database.logger.Debugf("Using database connection string: %s", connStr)

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		database.logger.Fatalf("CockroachDB Error: %s", err)
	}
	return db
}
