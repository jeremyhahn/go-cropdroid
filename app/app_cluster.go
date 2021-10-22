//go:build cluster
// +build cluster

package app

import (
	//#include <unistd.h>
	//#include <errno.h>
	"C"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jinzhu/gorm"
)
import "github.com/jeremyhahn/go-cropdroid/common"

/*
    	 Auth & Billing System              (gossip, dragonboat, cockroachdb - control plane) (cloud platform)
			 /            \
		(gossip)         (gossip)           <-- data center raft group (cockroach changefeeds create Enterprise state proposals)
		   /		         \
	   Cluster1            ClusterN         (ASGs & dragonboat)
	    /  \                 /  \
    Farm1  FarmN          Farm1  FarmN
     / \                          / \
 State Config                  State Config

	Gossip Data Model:
	- Timezone     (region)
	- Datacenter   (availability zone)  (nodes in a datacenter can form a Raft group, datacenters sync using gossip)

	Billing Raft Groups:
	- OLTP (cockroach)
	- orgs
	- users
	- roles
	- permissions
	- billing
	- licensing
	- provisioning
*/

const Name = "cropdroid"

type App struct {
	//configMutex         *sync.RWMutex
	ChannelIndex        state.ChannelIndex
	Config              *config.Server
	ConfigDir           string
	ConfigFile          string
	DeviceIndex         state.DeviceIndex
	DataStoreEngine     string
	DataStoreCDC        bool
	DataDir             string
	DebugFlag           bool
	DowngradeUser       string
	EnableRegistrations bool // WebRegistrations
	EnableDefaultFarm   bool
	GormDB              gormstore.GormDB
	GORM                *gorm.DB
	GORMInitParams      *gormstore.GormInitParams
	//GossipCluster       cluster.GossipCluster
	HomeDir  string
	Interval int
	KeyDir   string
	KeyPair  KeyPair
	Location *time.Location
	LogDir   string
	LogFile  string
	Logger   *logging.Logger
	Mailer   common.Mailer
	Mode     string
	Name     string
	NodeID   int
	//RaftCluster         cluster.RaftCluster
	RedirectHttpToHttps bool
	SSLFlag             bool
	WebPort             int
}

func NewApp() *App {
	return &App{Name: Name}
}

// func (this *App) GetConfig() *config.Server {
// 	this.configMutex.RLock()
// 	defer this.configMutex.RUnlock()
// 	return this.Config
// }

// func (this *App) SetConfig(serverConfig config.ServerConfig) {
// 	this.configMutex.Lock()
// 	defer this.configMutex.Unlock()
// 	this.Config = serverConfig.(*config.Server)
// }

func (this *App) ValidateConfig() {
	if this.Interval != 60 {
		this.Logger.Fatal("Interval must be 60 seconds")
	}
	if len(this.Config.Organizations[0].GetFarms()) <= 0 {
		this.Logger.Fatal("No farms configured, consider running 'cropdroid config --init' first")
	}
}

func (this *App) ValidateFreewareConfig() {
	this.ValidateConfig()
	if len(this.Config.Organizations) > 0 {
		if err := this.ValidateLicense(); err != nil {
			this.Logger.Fatal(err)
		}
	}
}

func (this *App) ValidateCloudConfig() {
	this.ValidateConfig()
	if len(this.Config.Organizations) >= 0 {
		this.Logger.Fatal("No organizations configured, consider running 'cropdroid config --init' first")
	}
}

func (this *App) ValidateLicense() error {
	unlicensed := "Unlicensed features configured. You must purchase a license to continue."
	if this.Config.License == nil {
		return fmt.Errorf("%s", unlicensed)
	}
	return nil
}

func (this *App) DropPrivileges() {
	if syscall.Getuid() == 0 {
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

func (this *App) InitLogFile(uid, gid int) *os.File {
	var f *os.File
	var err error
	f, err = os.OpenFile(this.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(this.LogFile)
	if err != nil {
		if os.IsNotExist(err) {
			_f, err2 := os.Create(this.LogFile)
			if err2 != nil {
				log.Fatal(err2)
			}
			f = _f
		}
		log.Fatal(err)
	}
	if uid == 0 {
		err = os.Chown(this.LogFile, uid, gid)
		if err != nil {
			log.Fatal(err)
		}
		if this.DebugFlag {
			err = os.Chmod(this.LogFile, 0777)
		} else {
			err = os.Chmod(this.LogFile, 0644)
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	return f
}

func (this *App) InitGormDB() *gorm.DB {
	this.GORM = this.NewGormDB()
	return this.GORM
}

func (this *App) NewGormDB() *gorm.DB {
	this.GormDB = gormstore.NewGormDB(this.Logger, this.GORMInitParams)
	return this.GormDB.Connect(false)
}
