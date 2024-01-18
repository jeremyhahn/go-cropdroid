package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var InitDB bool
var ShowConfigFlag bool
var CompressConfigFlag bool
var ConfigOutputFormat string

func init() {

	configCmd.PersistentFlags().BoolVarP(&InitDB, "init", "", false, "Initialize new application configuration")
	configCmd.PersistentFlags().BoolVarP(&ShowConfigFlag, "show", "", false, "Show the current configuration")
	configCmd.PersistentFlags().BoolVarP(&CompressConfigFlag, "compress", "", false, "Compress the current configuration")
	configCmd.PersistentFlags().StringVarP(&ConfigOutputFormat, "format", "", "yaml", "Output format [ yaml | json ]")

	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Prints the current configuration",
	Long:  `Displays the current system configuration`,
	Run: func(cmd *cobra.Command, args []string) {

		if InitDB {

			gormdb := gorm.NewGormDB(App.Logger, App.GORMInitParams)
			db := gormdb.Connect(true)
			db.LogMode(App.DebugFlag)

			idGenerator := util.NewIdGenerator(App.DataStoreEngine)

			switch App.Mode {
			case common.CONFIG_MODE_VIRTUAL:
				if err := gorm.NewGormInitializer(App.Logger, gormdb, idGenerator, App.Location,
					App.Mode).Initialize(App.EnableDefaultFarm); err != nil {
					log.Fatal(err)
				}
			// case common.MODE_CLOUD:
			// 	if err := gorm.NewGormCloudInitializer(App.Logger, App.GORM, App.Location).Initialize(); err != nil {
			// 		log.Fatal(err)
			// 	}
			default:
				App.Logger.Fatalf("Unsupported mode: %s", App.Mode)
			}

			log.Println("Database initialized")
			return
		}

		if ShowConfigFlag {

			// c := viper.AllSettings()
			// bs, err := yaml.Marshal(c)
			// if err != nil {
			// 	log.Fatalf("unable to marshal config to YAML: %v", err)
			// }
			// fmt.Printf("%+v\n", string(bs))

			if DataStoreEngine == "memory" || DataStoreEngine == "sqlite" ||
				DataStoreEngine == "postgres" || DataStoreEngine == "cockroach" {

				App.InitLogFile(os.Getuid(), os.Getgid())

				_, _, _, _, err := builder.NewGormConfigBuilder(App, DeviceDataStore,
					AppStateTTL, AppStateTick).Build()
				if err != nil {
					App.Logger.Fatal(err)
				}
				//spew.Dump(serverConfig)
				var data []byte
				if ConfigOutputFormat == "yaml" {
					data, err = yaml.Marshal(App)
				} else if ConfigOutputFormat == "json" {
					data, err = json.Marshal(App)
				} else {
					App.Logger.Fatal("Unsupported output format: %s", ConfigOutputFormat)
				}
				if err != nil {
					App.Logger.Fatal(err)
				}
				fmt.Printf("%+v\n", string(data))
			} else {
				fmt.Printf("%+v\n", App)
			}
		}

		if CompressConfigFlag {
			compressor := util.NewCompressor()
			var data []byte
			var err error
			if DataStoreEngine == "yaml" {
				data, err = yaml.Marshal(App)
			} else if DataStoreEngine == "json" {
				data, err = json.Marshal(App)
			} else {
				App.Logger.Fatalf("Unsupported datastore type: %s", DataStoreEngine)
			}
			if err != nil {
				App.Logger.Fatal(err)
			}
			gzipped, err := compressor.Zip(data)
			if err != nil {
				App.Logger.Fatal(err)
			}
			encoded := base64.StdEncoding.EncodeToString(gzipped)
			println(encoded)
		}

	},
}
