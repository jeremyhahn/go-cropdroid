package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/config"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	yamlstore "github.com/jeremyhahn/go-cropdroid/datastore/yaml"
	"github.com/jeremyhahn/go-cropdroid/mapper"
	logging "github.com/op/go-logging"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var App *app.App
var NodeID int
var AppStateTTL int
var AppStateTick int
var DebugFlag bool
var Interval int
var ConfigDir string
var DataDir string
var ConfigFile string
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

var DatastoreType string
var DatastoreUser string
var DatastorePass string
var DatastoreHost string
var DatastorePort int
var DatastoreCDC bool
var DatastoreCACert string
var DatastoreTlsKey string
var DatastoreTlsCert string

var DataStore string

//var supportedGormEngines = []string{"json", "yaml", "sqlite", "postgres", "cockroach"}
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
	//rootCmd.PersistentFlags().StringVarP(&DowngradeUser, "setuid", "", "www-data", "Root downgrade user/group")
	rootCmd.PersistentFlags().StringVarP(&DowngradeUser, "setuid", "", "root", "Root downgrade user/group")
	rootCmd.PersistentFlags().BoolVarP(&EnableRegistrations, "enable-registrations", "", false, "Allows user account registrations via API")

	rootCmd.PersistentFlags().StringVarP(&DatastoreType, "datastore", "", "memory", "Datastore type [ memory | sqlite | mysql | postgres | cockroach ]")
	rootCmd.PersistentFlags().StringVarP(&DatastoreUser, "datastore-user", "", "root", "Datastore username")
	rootCmd.PersistentFlags().StringVarP(&DatastorePass, "datastore-pass", "", "", "Datastore password")
	rootCmd.PersistentFlags().StringVarP(&DatastoreHost, "datastore-host", "", "localhost", "Datastore IP or hostname")
	rootCmd.PersistentFlags().IntVarP(&DatastorePort, "datastore-port", "", 26257, "Datastore listen port")
	rootCmd.PersistentFlags().BoolVarP(&DatastoreCDC, "datastore-cdc", "", false, "Enable database changefeed (Change Data Capture) real-time updates on supported tables")
	rootCmd.PersistentFlags().StringVarP(&DatastoreCACert, "datastore-ca-cert", "", "", "TLS Certificate Authority public key")
	rootCmd.PersistentFlags().StringVarP(&DatastoreTlsKey, "datastore-tls-key", "", "", "TLS key used to encrypt the database connection")
	rootCmd.PersistentFlags().StringVarP(&DatastoreTlsCert, "datastore-tls-cert", "", "", "TLS certificate used to encrypt the database connection")

	rootCmd.PersistentFlags().StringVarP(&DataStore, "data-store", "", "gorm", "Where to store historical device data [ gorm | redis ]")

	rootCmd.PersistentFlags().BoolVarP(&EnableDefaultFarm, "enable-default-farm", "", true, "Create a default farm on startup")

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
	App.NodeID = NodeID
	App.DatastoreType = DatastoreType
	App.DatastoreCDC = DatastoreCDC
	App.DebugFlag = DebugFlag
	App.Interval = Interval
	App.HomeDir = HomeDir
	App.DataDir = DataDir
	App.WebPort = WebPort
	App.SSLFlag = SSLFlag
	App.KeyDir = KeyDir
	App.RedirectHttpToHttps = RedirectHttpToHttps
	App.Mode = Mode
	App.DowngradeUser = DowngradeUser
	App.EnableRegistrations = EnableRegistrations
	App.EnableDefaultFarm = EnableDefaultFarm
	initLogger()
	initConfig()
	if App.DebugFlag {

		//listFiles()

		logging.SetLevel(logging.DEBUG, "")
		App.Logger.Debug("Starting logger in debug mode...")
		App.Logger.Debugf("ConfigFile: \t\t%s", ConfigFile)
		App.Logger.Debugf("DatastoreType: \t\t%s", DatastoreType)

		App.Logger.Debugf("DatastoreCACert: \t\t%s", DatastoreCACert)
		App.Logger.Debugf("DatastoreTlsKey: \t\t%s", DatastoreTlsKey)
		App.Logger.Debugf("DatastoreTlsCert: \t\t%s", DatastoreTlsCert)

		App.Logger.Debugf("HomeDir: \t\t%s", HomeDir)
		App.Logger.Debugf("ConfigDir: \t\t%s", ConfigDir)
		App.Logger.Debugf("DataDir: \t\t%s", DataDir)
		App.Logger.Debugf("LogDir: \t\t%s", LogDir)
		App.Logger.Debugf("KeyDir: \t\t%s", KeyDir)
		App.Logger.Debugf("Web service port: \t%d", WebPort)
		App.Logger.Debugf("Web service SSL: \t%t", SSLFlag)
		App.Logger.Debugf("Redirect HTTP: \t\t%t", RedirectHttpToHttps)
		App.Logger.Debugf("Timezone: \t\t%s", Timezone)
		App.Logger.Debugf("Mode: \t\t\t%s", Mode)
		App.Logger.Debugf("uid/gid: \t\t%s", DowngradeUser)
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	if SSLFlag && WebPort == 80 {
		App.WebPort = 443
	}
}

func listFiles() {
	files, err := ioutil.ReadDir("/cockroach/cockroach-certs/")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Printf("File: %s", file.Name())
	}

	libs, err := ioutil.ReadDir("/usr/local/lib/cockroach")
	if err != nil {
		log.Fatal(err)
	}
	for _, lib := range libs {
		fmt.Printf("Lib: %s", lib.Name())
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
	DatastoreType = strings.ToLower(DatastoreType)
	App.GORMInitParams = &gormstore.GormInitParams{
		AppMode:           Mode,
		DebugFlag:         DebugFlag,
		EnableDefaultFarm: EnableDefaultFarm,
		DataDir:           DataDir,
		Engine:            DatastoreType,
		Host:              DatastoreHost,
		Port:              DatastorePort,
		Username:          DatastoreUser,
		Password:          DatastorePass,
		CACert:            DatastoreCACert,
		TLSKey:            DatastoreTlsKey,
		TLSCert:           DatastoreTlsCert,
		DBName:            app.Name,
		Location:          App.Location}
	App.ConfigDir = ConfigDir
	configTypeSupported := func(configTypes []string) bool {
		for _, t := range configTypes {
			if DatastoreType == t {
				return true
			}
		}
		return false
	}(supportedGormEngines)
	if !configTypeSupported {
		log.Fatalf("Config type not supported: %s", DatastoreType)
	}
	if DatastoreType == "yaml" || DatastoreType == "json" {
		ConfigFile = fmt.Sprintf("%s/%s.%s", ConfigDir, "config", DatastoreType)
		App.ConfigFile = ConfigFile
		data, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			App.Logger.Fatal(err)
		}
		loadConfig(data)
		App.ValidateConfig()
	} else if DatastoreType == "sqlite" {
		App.GORMInitParams.DataDir = fmt.Sprintf("%s/db", HomeDir)
	}
}

func loadConfig(data []byte) {
	var serverConfig yamlstore.Server
	var err error
	if DatastoreType == "yaml" {
		App.Logger.Debugf("Deserializing YAML configuration (%s)", App.ConfigFile)
		err = yaml.Unmarshal(data, &serverConfig)
	} else if DatastoreType == "json" {
		App.Logger.Debug("Deserializing JSON configuration")
		err = json.Unmarshal(data, &serverConfig)
	} else {
		App.Logger.Fatal("Unexpected configuration type: %s", DatastoreType)
	}
	if err != nil {
		App.Logger.Fatal(err)
	}

	//App.Config = &serverConfig
	App.Logger.Debug(serverConfig)

	conf, err := mapper.NewConfigMapper().MapFromFileConfig(&serverConfig)
	if err != nil {
		App.Logger.Fatal(err)
	}
	App.Logger.Debug(conf)
	App.Config = conf.(*config.Server)
	App.Logger.Debug(App.Config)
}
