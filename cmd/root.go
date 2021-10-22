package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var App *app.App
var NodeID int
var AppStateTTL int
var AppStateTick int
var DebugFlag bool
var Interval int
var ConfigDir string
var DataDir string
var LogDir string
var LogFile string
var HomeDir string
var WebPort int
var ServerID int
var SSLFlag bool
var KeyDir string
var RedirectHttpToHttps bool
var Timezone string
var Mode string
var DowngradeUser string
var EnableRegistrations bool
var EnableDefaultFarm bool

var DataStore string
var DataStoreEngine string
var DataStoreUser string
var DataStorePass string
var DataStoreHost string
var DataStorePort int
var DataStoreCDC bool
var DataStoreCACert string
var DataStoreTlsKey string
var DataStoreTlsCert string

var DefaultRole string
var DefaultFarmPermission string

var supportedGormEngines = []string{"memory", "sqlite", "mysql", "postgres", "cockroach"}

var logFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortpkg}.%{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

var rootCmd = &cobra.Command{
	Use:   "cropdroid",
	Short: "Automated farming and cultivation",
	Long: `
  _______  ______  _____   _____  ______   ______  _____  _____ ______
  |       |_____/ |     | |_____] |     \ |_____/ |     |   |   |     \
  |_____  |    \_ |_____| |       |_____/ |    \_ |_____| __|__ |_____/

 Fully automate your hobby or commercial soil, hydroponics, or aeroponics
 indoor garden or outdoor farm with CropDroid. Gain real-time statistics,
 remote monitoring and administration, push notifications for critical alerts
 and sophisticated dosing algorithms with artificial intelligence to take the
 guess work out of feedings and ongoing maintenance. Configure one-time and/or
 recurring schedules to switch things on and off, program conditional logic based on
 sensor values, digital timers to control how long things stay on/off, and automate your
 own custom sequence of actions using programmable workflows. Enable cloud mode for easy
 world-wide access, store and archive your data, access powerful analytics, automate
 supply replenishment, access to social features, and an online database with valuable
 plant data and downloadable configurations to ease setup, maximize yields, and provide
 precise, reproducible grow conditions.

 CropDroid requires the use of hardware devices with sensors to monitor and manage your
 crop. A "room" device is used to control environment parameters for indoor grow
 rooms and greenhouses, including lights, temperature, humidity, and Co2. A "reservoir"
 device is used to manage water quality and flow while the "dosing" device
 allows precise amounts of nutrients and chemicals, and has the ability to act as a
 general purpose switching device. An "irrigation" device is used to individually monitor
 soil moisture on a pot by pot basis and hydrate them when necessary.Custom devices are
 also available to meet your specific requirements.

 Complete documentation is available at https://github.com/jeremyhahn/go-cropdroid`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//initApp()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
	TraverseChildren: true,
}

func init() {
	cobra.OnInitialize(initApp)

	wd, _ := os.Getwd()

	rootCmd.PersistentFlags().IntVarP(&NodeID, "node-id", "", 1, "Unique node identifier")
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().IntVarP(&ServerID, "sid", "", 1, "Unique Server ID")
	rootCmd.PersistentFlags().IntVarP(&Interval, "interval", "", 60, "Default poll interval (seconds)")
	rootCmd.PersistentFlags().StringVarP(&Mode, "mode", "", "virtual", "Service mode: [ virtual | edge ]")
	rootCmd.PersistentFlags().IntVarP(&AppStateTTL, "ttl", "", 0, "How long to keep farm in app state (seconds). 0 = never expire")
	rootCmd.PersistentFlags().IntVarP(&AppStateTick, "gctick", "", 3600, "How often to check farm store for expired entries")
	rootCmd.PersistentFlags().StringVarP(&HomeDir, "home", "", wd, "Program home directory") // doesnt work as system daemon if not wd (/)
	rootCmd.PersistentFlags().StringVarP(&DataDir, "data-dir", "", fmt.Sprintf("%s/db", wd), "Directory where database files are stored")
	rootCmd.PersistentFlags().StringVarP(&ConfigDir, "config-dir", "", "/etc/cropdroid", "Directory where configuration files are stored")
	rootCmd.PersistentFlags().StringVarP(&LogDir, "log-dir", "", "/var/log", "Logging directory")
	rootCmd.PersistentFlags().StringVarP(&LogFile, "log-file", "", "/var/log/cropdroid.log", "Application log file")
	rootCmd.PersistentFlags().IntVarP(&WebPort, "port", "", 80, "Web service port number")
	rootCmd.PersistentFlags().BoolVarP(&SSLFlag, "ssl", "", true, "Enable web service SSL / TLS")
	rootCmd.PersistentFlags().StringVarP(&KeyDir, "keys", "", fmt.Sprintf("%s/keys", wd), "Directory where key files are stored")
	rootCmd.PersistentFlags().BoolVarP(&RedirectHttpToHttps, "redirect-http-https", "", false, "Redirect HTTP to HTTPS")
	rootCmd.PersistentFlags().StringVarP(&Timezone, "timezone", "", "America/New_York", "Local time zone")
	rootCmd.PersistentFlags().StringVarP(&DowngradeUser, "setuid", "", "root", "Root downgrade user/group")
	rootCmd.PersistentFlags().BoolVarP(&EnableRegistrations, "enable-registrations", "", false, "Allows user account registrations via API")

	rootCmd.PersistentFlags().StringVarP(&DataStoreEngine, "datastore", "", "memory", "Data store type [ memory | sqlite | mysql | postgres | cockroach ]")
	rootCmd.PersistentFlags().StringVarP(&DataStoreUser, "datastore-user", "", "root", "Data store username")
	rootCmd.PersistentFlags().StringVarP(&DataStorePass, "datastore-pass", "", "", "Data store password")
	rootCmd.PersistentFlags().StringVarP(&DataStoreHost, "datastore-host", "", "localhost", "Data store IP or hostname")
	rootCmd.PersistentFlags().IntVarP(&DataStorePort, "datastore-port", "", 26257, "Data store listen port")
	rootCmd.PersistentFlags().BoolVarP(&DataStoreCDC, "datastore-cdc", "", false, "Enable database changefeed (Change Data Capture) real-time updates on supported tables")
	rootCmd.PersistentFlags().StringVarP(&DataStoreCACert, "datastore-ca-cert", "", "", "TLS Certificate Authority public key")
	rootCmd.PersistentFlags().StringVarP(&DataStoreTlsKey, "datastore-tls-key", "", "", "TLS key used to encrypt the database connection")
	rootCmd.PersistentFlags().StringVarP(&DataStoreTlsCert, "datastore-tls-cert", "", "", "TLS certificate used to encrypt the database connection")

	rootCmd.PersistentFlags().StringVarP(&DataStore, "data-store", "", "gorm", "Where to store historical device data [ gorm | redis ]")

	rootCmd.PersistentFlags().BoolVarP(&EnableDefaultFarm, "enable-default-farm", "", true, "Create a default farm on startup")
	rootCmd.PersistentFlags().StringVarP(&DefaultRole, "default-role", "", "admin", "Default role to assign to newly registered users [ admin | cultivator | analyst ]")
	rootCmd.PersistentFlags().StringVarP(&DefaultFarmPermission, "default-permission", "", "all", "Default permission given to newly registered users to access existing farms [ all | owner | none ]")

	viper.BindPFlags(rootCmd.PersistentFlags())

	if runtime.GOOS == "darwin" {
		signal.Ignore(syscall.Signal(0xd))
	}
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func initApp() {
	location, err := time.LoadLocation(Timezone)
	if err != nil {
		log.Fatalf("Unable to parse default timezone %s: %s", location, err)
	}
	App.Location = location
	App.DebugFlag = viper.GetBool("debug")
	App.HomeDir = viper.GetString("home")
	App.KeyDir = viper.GetString("keys")
	App.DataStoreEngine = viper.GetString("datastore")
	initLogger()
	initConfig()
	if App.DebugFlag {
		//listFiles()
		logging.SetLevel(logging.DEBUG, "")
		App.Logger.Debug("Starting logger in debug mode...")
		for k, v := range viper.AllSettings() {
			App.Logger.Debugf("%s: %+v", k, v)
		}
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	if viper.GetBool("ssl") && viper.GetInt("port") == 80 {
		App.Config.WebPort = 443
	}
}

func initLogger() {
	App.LogDir = LogDir
	App.LogFile = LogFile
	f := App.InitLogFile(os.Getuid(), os.Getgid())
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logfile := logging.NewLogBackend(f, "", log.Lshortfile)
	logFormatter := logging.NewBackendFormatter(logfile, logFormat)
	//syslog, _ := logging.NewSyslogBackend(appName)
	backends := logging.MultiLogger(stdout, logFormatter)
	logging.SetBackend(backends)
	if App.DebugFlag {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.ERROR, "")
	}
	App.Logger = logging.MustGetLogger(app.Name)
}

func initConfig() {
	datastoreEngine := viper.GetString("datastore")
	App.GORMInitParams = &gormstore.GormInitParams{
		AppMode:           viper.GetString("mode"),
		DebugFlag:         viper.GetBool("debug"),
		EnableDefaultFarm: viper.GetBool("enable-default-farm"),
		DataDir:           viper.GetString("data-dir"),
		Engine:            datastoreEngine,
		Host:              viper.GetString("datastore-host"),
		Port:              viper.GetInt("datastore-port"),
		Username:          viper.GetString("datastore-user"),
		Password:          viper.GetString("datastore-pass"),
		CACert:            viper.GetString("datastore-ca-cert"),
		TLSKey:            viper.GetString("datastore-tls-key"),
		TLSCert:           viper.GetString("datastore-tls-cert"),
		DBName:            app.Name,
		Location:          App.Location}

	configTypeSupported := false
	for _, t := range supportedGormEngines {
		if datastoreEngine == t {
			configTypeSupported = true
			break
		}
	}
	if !configTypeSupported {
		log.Fatalf("Config type not supported: %s", datastoreEngine)
	}

	if datastoreEngine == "sqlite" {
		App.GORMInitParams.DataDir = fmt.Sprintf("%s/db", HomeDir)
	}

	App.Config = &config.Server{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(ConfigDir)
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", App.Name))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s/", App.Name))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		App.Logger.Errorf("%s", err)
	}

	viper.Unmarshal(&App.Config)

	App.Config.NodeID = viper.GetInt("node-id")
	App.Config.DefaultRole = viper.GetString("default-role")
	App.Config.DefaultPermission = viper.GetString("default-permission")
	App.Config.DataStoreEngine = viper.GetString("datastore")
	App.Config.DataStoreCDC = viper.GetBool("datastore-cdc")
	App.Config.Interval = viper.GetInt("interval")
	App.Config.DataDir = viper.GetString("data-dir")
	App.Config.WebPort = viper.GetInt("port")
	App.Config.SSLFlag = viper.GetBool("ssl")
	App.Config.RedirectHttpToHttps = viper.GetBool("redirect-http-https")
	App.Config.Mode = viper.GetString("mode")
	App.Config.DowngradeUser = viper.GetString("setuid")
	App.Config.EnableRegistrations = viper.GetBool("enable-registrations")
	App.Config.EnableDefaultFarm = viper.GetBool("enable-default-farm")

	App.Logger.Debugf("%+v", App.Config)

	//App.ValidateConfig()
}

// func listFiles() {
// 	files, err := ioutil.ReadDir("/cockroach/cockroach-certs/")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, file := range files {
// 		fmt.Printf("File: %s", file.Name())
// 	}

// 	libs, err := ioutil.ReadDir("/usr/local/lib/cockroach")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, lib := range libs {
// 		fmt.Printf("Lib: %s", lib.Name())
// 	}
// }
