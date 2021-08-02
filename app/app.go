// +build !cluster

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
	"sync"
	"syscall"
	"time"

	logging "github.com/op/go-logging"

	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jinzhu/gorm"
)

const Name = "cropdroid"

type App struct {
	configMutex         *sync.RWMutex
	ChannelIndex        state.ChannelIndex
	Config              *config.Server
	ConfigDir           string
	ConfigFile          string
	DeviceIndex         state.DeviceIndex
	DatastoreType       string
	DatastoreCDC        bool
	DataDir             string
	DebugFlag           bool
	DowngradeUser       string
	EnableRegistrations bool // WebRegistrations
	GormDB              gormstore.GormDB
	GORM                *gorm.DB
	GORMInitParams      *gormstore.GormInitParams
	HomeDir             string
	Interval            int
	KeyDir              string
	KeyPair             KeyPair
	Location            *time.Location
	LogDir              string
	LogFile             string
	Logger              *logging.Logger
	Mode                string
	Name                string
	NodeID              int
	RedirectHttpToHttps bool
	SSLFlag             bool
	WebPort             int
}

func NewApp() *App {
	return &App{Name: Name}
}

func (this *App) GetConfig() *config.Server {
	this.configMutex.RLock()
	defer this.configMutex.RUnlock()
	return this.Config
}

func (this *App) SetConfig(serverConfig config.ServerConfig) {
	this.configMutex.Lock()
	defer this.configMutex.Unlock()
	this.Config = serverConfig.(*config.Server)
}

func (this *App) ValidateConfig() {
	if this.Interval < 10 {
		this.Logger.Fatal("Interval must be at least 10 seconds")
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

func (this *App) InitGormDB() *gorm.DB {
	this.GORM = this.NewGormDB()
	return this.GORM
}

func (this *App) NewGormDB() *gorm.DB {
	return gormstore.NewGormDB(this.Logger, this.GORMInitParams).Connect(false)
}
