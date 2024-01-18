package app

import (
	//#include <unistd.h>
	//#include <errno.h>
	"C"
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/util"
)

//     	 Auth & Billing System              (gossip, dragonboat, cockroachdb - control plane) (cloud platform)
// 			 /            \
// 		(gossip)         (gossip)           <-- data center raft group (cockroach changefeeds create Enterprise state proposals)
// 		   /		         \
// 	   Cluster1            ClusterN         (ASGs & dragonboat)
// 	    /  \                 /  \
//     Farm1  FarmN          Farm1  FarmN
//      / \                      / \
//  State Config              State Config

// 	Gossip Data Model:
// 	- Timezone     (region)
// 	- Datacenter   (availability zone)  (nodes in a datacenter can form a Raft group, datacenters sync using gossip)

// 	Billing Raft Groups:
// 	- OLTP (cockroach)
// 	- orgs
// 	- users
// 	- roles
// 	- permissions
// 	- billing
// 	- licensing
// 	- provisioning

type App struct {
	ClusterID               uint64 `yaml:"clusterId" json:"cluster_id" mapstructure:"cluster_id"`
	DebugFlag               bool   `yaml:"debug" json:"debug" mapstructure:"debug"`
	DataDir                 string `yaml:"datadir" json:"datadir" mapstructure:"datadir"`
	DataStoreEngine         string `yaml:"datastore" json:"datastore" mapstructure:"datastore"`
	DataStoreCDC            bool   `yaml:"datastore_cdc" json:"datastore_cdc" mapstructure:"datastore_cdc"`
	DefaultRole             string `yaml:"default_role" json:"default_role" mapstructure:"default_role"`
	DefaultPermission       string `yaml:"default_permission" json:"default_permission" mapstructure:"default_permission"`
	DefaultConsistencyLevel int    `yaml:"default_consistency_level" json:"default_consistency_level" mapstructure:"default_consistency_level"`
	DefaultConfigStoreType  int    `yaml:"default_config_store" json:"default_config_store" mapstructure:"default_config_store"`
	DefaultStateStoreType   int    `yaml:"default_state_store" json:"default_state_store" mapstructure:"default_state_store"`
	DefaultDataStoreType    int    `yaml:"default_data_store" json:"default_data_store" mapstructure:"default_data_store"`
	DowngradeUser           string `yaml:"www_user" json:"www_user" mapstructure:"www_user"`
	EnableDefaultFarm       bool   `yaml:"enable_default_farm" json:"enable_default_farm" mapstructure:"enable_default_farm"`
	EnableRegistrations     bool   `yaml:"enable_registrations" json:"enable_registrations" mapstructure:"enable_registrations"`
	//Farms               []config.Farm `gorm:"-" yaml:"farms" json:"farms"`
	//GormDB         gormstore.GormDB
	GORMInitParams *gormstore.GormInitParams
	HomeDir        string `yaml:"home_dir" json:"home_dir" mapstructure:"home_dir"`
	IdGenerator    util.IdGenerator
	IdSetter       util.IdSetter
	Interval       int    `yaml:"interval" json:"interval" mapstructure:"interval"`
	KeyDir         string `yaml:"key_dir" json:"key_dir" mapstructure:"key_dir"`
	KeyPair        KeyPair
	LicenseBlob    string          `yaml:"license" json:"license" mapstructure:"license"`
	License        *config.License `yaml:"-" json:"-" mapstructure:"-"`
	Location       *time.Location
	LogDir         string `yaml:"log_dir" json:"log_dir" mapstructure:"log_dir"`
	LogFile        string `yaml:"log_file" json:"log_file" mapstructure:"log_file"`
	Logger         *logging.Logger
	Mailer         common.Mailer
	Mode           string `yaml:"mode" json:"mode" mapstructure:"mode"`
	Name           string
	NodeID         int `yaml:"node_id" json:"node_id" mapstructure:"node_id"`
	//Organizations       []config.Organization `yaml:"organizations" json:"organizations" mapstructure:"organizations"`
	RedirectHttpToHttps bool         `yaml:"redirect_http_https" json:"redirect_http_https" mapstructure:"redirect_http_https"`
	Smtp                *config.Smtp `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	SSLFlag             bool         `yaml:"ssl" json:"ssl" mapstructure:"ssl"`
	StateTTL            int          `yaml:"state_ttl" json:"state_ttl" mapstructure:"state_ttl"`
	StateTick           int          `yaml:"state_tick" json:"state_tick" mapstructure:"state_tick"`
	Timezone            string       `yaml:"timezone" json:"timezone" mapstructure:"timezone"`
	WebPort             int          `yaml:"port" json:"port" mapstructure:"port"`
}

func NewApp() *App {
	return &App{Name: "cropdroid"}
}

// func (this *App) InitGormDB() *gorm.DB {
// 	this.GORM = this.NewGormDB()
// 	return this.GORM
// }

// func (this *App) NewGormDB() *gorm.DB {
// 	this.GormDB = gormstore.NewGormDB(this.Logger, this.GORMInitParams)
// 	return this.GormDB.Connect(false)
// }

func (this *App) InitLogFile(uid, gid int) *os.File {
	var logFile string
	if this.LogFile == "" {
		logFile = fmt.Sprintf("%s/cropdroid.log", this.LogDir)
	} else {
		logFile = this.LogFile
	}
	var f *os.File
	var err error
	f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			_f, err2 := os.Create(logFile)
			if err2 != nil {
				log.Fatal(err2)
			}
			f = _f
		}
		log.Fatal(err)
	}
	if uid == 0 {
		err = os.Chown(logFile, uid, gid)
		if err != nil {
			log.Fatal(err)
		}
		if this.DebugFlag {
			err = os.Chmod(logFile, 0777)
		} else {
			err = os.Chmod(logFile, 0644)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	return f
}

func (this *App) DropPrivileges() {
	if runtime.GOOS != "linux" {
		return
	}
	if syscall.Getuid() == 0 && this.DowngradeUser != "root" {
		this.Logger.Debugf("Running as root, downgrading to user %s", this.DowngradeUser)
		user, err := user.Lookup(this.DowngradeUser)
		if err != nil {
			log.Fatalf("setuid %s user not found! Error: %s", this.DowngradeUser, err)
		}
		if err = syscall.Chdir(this.HomeDir); err != nil {
			this.Logger.Fatalf("Unable to chdir: error=%s", err)
		}
		uid, err := strconv.ParseInt(user.Uid, 10, 32)
		if err != nil {
			this.Logger.Fatalf("Unable to parse UID: %s", err)
		}
		gid, err := strconv.ParseInt(user.Gid, 10, 32)
		if err != nil {
			this.Logger.Fatalf("Unable to parse GID: %s", err)
		}
		cerr, errno := C.setgid(C.__gid_t(gid))
		if cerr != 0 {
			this.Logger.Fatalf("Unable to setgid: errno=%d: message=%s", errno, cerr)
		}
		cerr, errno = C.setuid(C.__uid_t(uid))
		if cerr != 0 {
			this.Logger.Fatalf("Unable to setuid: errno=%d: message=%s", errno, cerr)
		}
		this.InitLogFile(int(uid), int(gid))
	}
}
