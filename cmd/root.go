package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var App *app.App
var DebugFlag bool
var ConfigDir string
var LogDir string

var DatabaseInit bool
var DeviceDataStore string
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

var rootCmd = &cobra.Command{
	Use:   "cropdroid",
	Short: "Automated agriculture and local farmers market",
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
	 soil moisture on a pot by pot basis and hydrate them when necessary. Custom devices are
	 also available via Professional Services to meet your specific requirements, or you can
	 build your own and seamlessly integrate it with the rest of the CropDroid ecosystem.
	
	 Complete documentation is available at https://github.com/jeremyhahn/go-cropdroid`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//initApp()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
	TraverseChildren: true,
}

func init() {

	App = app.NewApp()

	cobra.OnInitialize(func() {
		App.Init(&app.AppInitParams{
			Debug:     DebugFlag,
			LogDir:    LogDir,
			ConfigDir: ConfigDir})
	})

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("unable to get current working directory")
		os.Exit(1)
	}

	// Required options to bootstrap the app
	rootCmd.PersistentFlags().BoolVarP(&DebugFlag, "debug", "", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&LogDir, "log-dir", "", "/var/log", "Logging directory")
	rootCmd.PersistentFlags().StringVarP(&ConfigDir, "config-dir", "", "/etc/cropdroid", "Directory where configuration files are stored")

	// Global options
	rootCmd.PersistentFlags().StringVarP(&App.Domain, "domain", "d", "localhost", "Domain name used to establish the CA and access web services")
	rootCmd.PersistentFlags().Uint64VarP(&App.NodeID, "node-id", "", 1, "Unique node identifier")
	rootCmd.PersistentFlags().IntVarP(&App.Interval, "interval", "", 60, "Default poll interval (seconds)")
	rootCmd.PersistentFlags().StringVarP(&App.Mode, "mode", "", "virtual", "Service mode: [ virtual | server | cloud | maintenance ]")
	rootCmd.PersistentFlags().StringVarP(&App.HomeDir, "home", "", wd, "Program home directory") // doesnt work as system daemon if not wd (/)
	rootCmd.PersistentFlags().StringVarP(&App.DataDir, "data-dir", "", fmt.Sprintf("%s/db", wd), "Directory where database files are stored")
	rootCmd.PersistentFlags().StringVarP(&App.Timezone, "timezone", "", "America/New_York", "Local time zone")
	rootCmd.PersistentFlags().StringVarP(&App.DowngradeUser, "setuid", "", "root", "Root downgrade user/group")
	rootCmd.PersistentFlags().BoolVarP(&App.EnableDefaultFarm, "enable-default-farm", "", false, "Create a default farm on startup")
	rootCmd.PersistentFlags().StringVarP(&App.DefaultRole, "default-role", "", "admin", "Default role to assign to newly registered users [ admin | cultivator | analyst ]")
	rootCmd.PersistentFlags().StringVarP(&App.DefaultPermission, "default-permission", "", "all", "Default permission given to newly registered users to access existing farms [ all | owner | none ]")

	// State store options
	rootCmd.PersistentFlags().IntVarP(&App.StateTTL, "state-ttl", "", 0, "How long to keep farm in app state (seconds). 0 = never expire")
	rootCmd.PersistentFlags().IntVarP(&App.StateTick, "state-tick", "", 3600, "How often to check farm store for expired entries")

	// Data store options
	rootCmd.PersistentFlags().StringVarP(&DeviceDataStore, "data-store", "", "gorm", "Where to store historical device data [ gorm | redis ]")

	// Web service options
	rootCmd.PersistentFlags().IntVarP(&App.WebPort, "web-port", "", 80, "Web service port number")
	rootCmd.PersistentFlags().IntVarP(&App.WebTlsPort, "web-tls-port", "", 443, "Web service TLS port number")
	rootCmd.PersistentFlags().IntVarP(&App.JwtExpiration, "jwt-expiration", "", 525600, "JWT expiration (minutes). Default 1 year")
	rootCmd.PersistentFlags().StringVarP(&App.CertDir, "cert-dir", "", fmt.Sprintf("%s/db/certs", wd), "Directory where key files are stored")
	rootCmd.PersistentFlags().BoolVarP(&App.RedirectHttpToHttps, "redirect-http-https", "", false, "Redirect HTTP to HTTPS")
	rootCmd.PersistentFlags().BoolVarP(&App.EnableRegistrations, "enable-registrations", "", false, "Allows user account registrations via API")

	// Database options
	rootCmd.PersistentFlags().BoolVarP(&DatabaseInit, "init", "", false, "Initialize an empty database with a default user and optional farm")
	rootCmd.PersistentFlags().StringVarP(&DataStoreEngine, "datastore", "", "memory", "Data store type [ memory | sqlite | mysql | postgres | cockroach ]")
	rootCmd.PersistentFlags().StringVarP(&DataStoreUser, "datastore-user", "", "root", "Data store username")
	rootCmd.PersistentFlags().StringVarP(&DataStorePass, "datastore-pass", "", "", "Data store password")
	rootCmd.PersistentFlags().StringVarP(&DataStoreHost, "datastore-host", "", "localhost", "Data store IP or hostname")
	rootCmd.PersistentFlags().IntVarP(&DataStorePort, "datastore-port", "", 26257, "Data store listen port")
	rootCmd.PersistentFlags().BoolVarP(&DataStoreCDC, "datastore-cdc", "", false, "Enable database changefeed (Change Data Capture) real-time updates on supported tables")
	rootCmd.PersistentFlags().StringVarP(&DataStoreCACert, "datastore-ca-cert", "", "", "TLS Certificate Authority public key")
	rootCmd.PersistentFlags().StringVarP(&DataStoreTlsKey, "datastore-tls-key", "", "", "TLS key used to encrypt the database connection")
	rootCmd.PersistentFlags().StringVarP(&DataStoreTlsCert, "datastore-tls-cert", "", "", "TLS certificate used to encrypt the database connection")

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
