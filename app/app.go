package app

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
	"time"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/pki/ca"
	"github.com/jeremyhahn/go-cropdroid/util"
)

//     	   cloud.cropdroid.com              (cloud: seed nodes, gossip, dragonboat, OLTP, OLAP, commercial control plane - auth, billing, & licensing)
// 			 /             \
// 		(gossip)         (gossip)           (region: data center, edge data aggregation, network overlay)
// 		   /		         \
// 	   Cluster1            ClusterN         (edge: data plane, gossip, dragonboat, raft and/or OLTP)
// 	     / \                 / \
//   Farm1  FarmN        Farm1  FarmN       (farm: collection of IoT devices)
//       / \                  / \
//  State & Config       State & Config     (device: user owned IoT devices with configurations and state telemetry)

var supportedDatastores = []string{"sqlite-memory", "sqlite", "mysql", "postgres"}

type App struct {
	CA                      ca.CertificateAuthority    `yaml:"-" json:"-" mapstructure:"-"`
	CAConfig                *ca.Config                 `yaml:"certificate-authority" json:"certificate_authority" mapstructure:"certificate-authority"`
	CertDir                 string                     `yaml:"cert-dir" json:"cert_dir" mapstructure:"cert-dir"`
	ClusterID               uint64                     `yaml:"cluster-id" json:"cluster_id" mapstructure:"cluster-id"`
	ConfigDir               string                     `yaml:"config-dir" json:"config_dir" mapstructure:"config-dir"`
	DatabaseInit            bool                       `yaml:"database-init" json:"database_init" mapstructure:"database-init"`
	DebugFlag               bool                       `yaml:"debug" json:"debug" mapstructure:"debug"`
	DataDir                 string                     `yaml:"data-dir" json:"data_dir" mapstructure:"data-dir"`
	DataStoreEngine         string                     `yaml:"datastore" json:"datastore" mapstructure:"datastore"`
	DefaultRole             string                     `yaml:"default-role" json:"default_role" mapstructure:"default-role"`
	DefaultPermission       string                     `yaml:"default-permission" json:"default_permission" mapstructure:"default-permission"`
	DefaultConsistencyLevel int                        `yaml:"default-consistency-level" json:"default_consistency_level" mapstructure:"default-onsistency-level"`
	DefaultConfigStoreType  int                        `yaml:"default-config-store" json:"default_config_store" mapstructure:"default-config-store"`
	DefaultStateStoreType   int                        `yaml:"default-state-store" json:"default_state_store" mapstructure:"default-state-store"`
	DefaultDataStoreType    int                        `yaml:"default-data-store" json:"default_data_store" mapstructure:"default-data-store"`
	Domain                  string                     `yaml:"domain" json:"domain" mapstructure:"domain"`
	DowngradeUser           string                     `yaml:"www-user" json:"www_user" mapstructure:"www-user"`
	EnableDefaultFarm       bool                       `yaml:"enable-default=farm" json:"enable_default_farm" mapstructure:"enable-default-farm"`
	EnableRegistrations     bool                       `yaml:"enable-registrations" json:"enable_registrations" mapstructure:"enable-registrations"`
	GORMInitParams          *gormstore.GormInitParams  `yaml:"-" json:"-" mapstructure:"-"`
	HomeDir                 string                     `yaml:"home-dir" json:"home-dir" mapstructure:"home-dir"`
	IdGenerator             util.IdGenerator           `yaml:"-" json:"-" mapstructure:"-"`
	IdSetter                util.IdSetter              `yaml:"-" json:"-" mapstructure:"-"`
	Interval                int                        `yaml:"interval" json:"interval" mapstructure:"interval"`
	JwtExpiration           int                        `yaml:"jwt-expiration" json:"jwt_expiration" mapstructure:"jwt-expiration"`
	LicenseBlob             string                     `yaml:"license" json:"license" mapstructure:"license"`
	ServerLicense           *config.ServerLicense      `yaml:"-" json:"-" mapstructure:"-"`
	Location                *time.Location             `yaml:"-" json:"-" mapstructure:"-"`
	LogDir                  string                     `yaml:"log-dir" json:"log_dir" mapstructure:"log-dir"`
	LogFile                 string                     `yaml:"log-file" json:"log_file" mapstructure:"log-file"`
	Logger                  *logging.Logger            `yaml:"-" json:"-" mapstructure:"-"`
	Mode                    string                     `yaml:"mode" json:"mode" mapstructure:"mode"`
	Name                    string                     `yaml:"-" json:"-" mapstructure:"-"`
	NodeID                  uint64                     `yaml:"node-id" json:"node_id" mapstructure:"node-id"`
	PasswordHasherParams    *util.PasswordHasherParams `yaml:"argon2" json:"argon2" mapstructure:"argon2"`
	RedirectHttpToHttps     bool                       `yaml:"redirect-http-https" json:"redirect_http_https" mapstructure:"redirect-http-https"`
	ShutdownChan            chan bool                  `yaml:"-" json:"-" mapstructure:"-"`
	Smtp                    *config.SmtpStruct         `yaml:"smtp" json:"smtp" mapstructure:"smtp"`
	Stripe                  *config.Stripe             `yaml:"stripe" json:"stripe" mapstructure:"stripe"`
	StateTTL                int                        `yaml:"state-ttl" json:"state_ttl" mapstructure:"state-ttl"`
	StateTick               int                        `yaml:"state-tick" json:"state_tick" mapstructure:"state-tick"`
	Timezone                string                     `yaml:"timezone" json:"timezone" mapstructure:"timezone"`
	WebPort                 int                        `yaml:"web-port" json:"web_port" mapstructure:"web-port"`
	WebTlsPort              int                        `yaml:"web-tls-port" json:"web_tls_port" mapstructure:"web-tls-port"`
}

type AppInitParams struct {
	Debug     bool
	LogDir    string
	ConfigDir string
}

func NewApp() *App {
	return &App{
		Name:         "cropdroid",
		ShutdownChan: make(chan bool, 1)}
}

func (app *App) Init(initParams *AppInitParams) {
	app.DebugFlag = initParams.Debug
	app.LogDir = initParams.LogDir
	app.ConfigDir = initParams.ConfigDir
	app.initLogger()
	app.initConfig()
	app.initCA()
	location, err := time.LoadLocation(app.Timezone)
	if err != nil {
		log.Fatalf("invalid timezone %s: %s", location, err)
	}
	app.Location = location
}

func (app *App) initLogger() {
	logFormat := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortpkg}.%{longfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
	)
	f := app.InitLogFile(os.Getuid(), os.Getgid())
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logfile := logging.NewLogBackend(f, "", log.Lshortfile)
	logFormatter := logging.NewBackendFormatter(logfile, logFormat)
	//syslog, - := logging.NewSyslogBackend(appName)
	backends := logging.MultiLogger(stdout, logFormatter)
	logging.SetBackend(backends)
	if app.DebugFlag {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.ERROR, "")
	}
	app.Logger = logging.MustGetLogger(app.Name)
	if app.DebugFlag {
		logging.SetLevel(logging.DEBUG, "")
		app.Logger.Debug("Starting logger in debug mode...")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
}

func (app *App) initConfig() {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(app.ConfigDir)
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", app.Name))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s/", app.Name))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		app.Logger.Errorf("%s", err)
	}

	app.Logger.Debugf("Using configuration file: %s",
		viper.GetViper().ConfigFileUsed())

	viper.Unmarshal(&app)

	// app.DebugFlag = viper.GetBool("debug")
	// app.HomeDir = viper.GetString("home")
	// app.CertDir = viper.GetString("cert-dir")
	app.DataStoreEngine = viper.GetString("datastore")
	app.IdGenerator = util.NewIdGenerator(app.DataStoreEngine)

	app.GORMInitParams = &gormstore.GormInitParams{
		AppMode:           viper.GetString("mode"),
		DebugFlag:         viper.GetBool("debug"),
		EnableDefaultFarm: viper.GetBool("enable-default-farm"),
		DataDir:           viper.GetString("data-dir"),
		Engine:            app.DataStoreEngine,
		Host:              viper.GetString("datastore-host"),
		Port:              viper.GetInt("datastore-port"),
		Username:          viper.GetString("datastore-user"),
		Password:          viper.GetString("datastore-pass"),
		CACert:            viper.GetString("datastore-ca-cert"),
		TLSKey:            viper.GetString("datastore-tls-key"),
		TLSCert:           viper.GetString("datastore-tls-cert"),
		DBName:            app.Name,
		Location:          app.Location}

	configTypeSupported := false
	for _, t := range supportedDatastores {
		if app.DataStoreEngine == t {
			configTypeSupported = true
			break
		}
	}
	if !configTypeSupported {
		log.Fatalf("datastore engine not supported: %s", app.DataStoreEngine)
	}

	// app.Timezone = viper.GetString("timezone")
	// app.NodeID = viper.GetInt("node-id")
	// app.DefaultRole = viper.GetString("default-role")
	// app.DefaultPermission = viper.GetString("default-permission")
	// app.DataStoreEngine = viper.GetString("datastore")
	// app.Interval = viper.GetInt("interval")
	// app.DataDir = viper.GetString("data-dir")
	// app.WebPort = viper.GetInt("port")
	// app.SSLFlag = viper.GetBool("ssl")
	// app.RedirectHttpToHttps = viper.GetBool("redirect-http-https")
	// app.Mode = viper.GetString("mode")
	// app.DowngradeUser = viper.GetString("setuid")
	// app.EnableRegistrations = viper.GetBool("enable-registrations")
	// app.EnableDefaultFarm = viper.GetBool("enable-default-farm")

	if viper.Get("argon2") == nil {
		app.PasswordHasherParams = &util.PasswordHasherParams{
			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32}
	}

	yamlConfig, err := yaml.Marshal(app)
	if err != nil {
		app.Logger.Fatalf("%s", yamlConfig)
	}

	// for k, v := range viper.AllSettings() {
	// 	app.Logger.Debugf("%s: %+v", k, v)
	// }
	app.Logger.Debugf("%s", yamlConfig)

	//app.ValidateConfig()
}

// Initializes the Certificate Authority. If this is the first time the CA is
// being initialized, new root CA private and public certificates are created
// and a web server certificate is issued for the configured domain. If the CA
// or web server certificate already exist, they are loaded from the cert store.
func (app *App) initCA() {
	CA, err := ca.NewCertificateAuthority(app.Logger, app.CertDir, app.CAConfig)
	if err != nil {
		app.Logger.Fatal(err)
	}
	CA.Init()
	_, err = CA.PEM(app.Domain)
	if err != nil {
		if err == ca.ErrCertNotFound {
			CA.IssueCertificate(ca.CertificateRequest{
				Valid:   365, // days
				Subject: app.CAConfig.Identity.Subject,
				SANS:    app.CAConfig.Identity.SANS,
			})
		} else {
			app.Logger.Fatal(err)
		}
	}
	app.CA = CA
}

// Initializes the application log file
func (app *App) InitLogFile(uid, gid int) *os.File {
	var logFile string
	if app.LogFile == "" {
		logFile = fmt.Sprintf("%s/%s.log", app.LogDir, app.Name)
		app.LogFile = logFile
	} else {
		logFile = app.LogFile
	}
	var f *os.File
	var err error
	f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	exists, err := util.FileExists(logFile)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		_, err = os.Create(logFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	if uid == 0 {
		err = os.Chown(logFile, uid, gid)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Chmod(logFile, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	return f
}

func (app *App) DropPrivileges() {
	if runtime.GOOS != "linux" {
		return
	}
	if syscall.Getuid() == 0 && app.DowngradeUser != "root" {
		app.Logger.Debugf("Running as root, downgrading to user %s", app.DowngradeUser)
		user, err := user.Lookup(app.DowngradeUser)
		if err != nil {
			log.Fatalf("setuid %s user not found! Error: %s", app.DowngradeUser, err)
		}
		if err = syscall.Chdir(app.HomeDir); err != nil {
			app.Logger.Fatalf("Unable to chdir: error=%s", err)
		}
		uid, err := strconv.ParseInt(user.Uid, 10, 32)
		if err != nil {
			app.Logger.Fatalf("Unable to parse UID: %s", err)
		}
		gid, err := strconv.ParseInt(user.Gid, 10, 32)
		if err != nil {
			app.Logger.Fatalf("Unable to parse GID: %s", err)
		}
		cerr := syscall.Setgid(int(gid))
		if cerr != nil {
			app.Logger.Fatalf("Unable to setgid: message=%s", cerr)
		}
		cerr = syscall.Setuid(int(uid))
		if cerr != nil {
			app.Logger.Fatalf("Unable to setuid: message=%s", cerr)
		}
		app.InitLogFile(int(uid), int(gid))
	}
}
