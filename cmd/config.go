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

			gormDB := gorm.NewGormDB(App.Logger, App.GORMInitParams)
			db := gormDB.Connect(true)
			gormDB.Create()

			switch App.Mode {
			case common.MODE_STANDALONE, "virtual":
				if err := gorm.NewGormInitializer(App.Logger, db, App.Location).Initialize(); err != nil {
					log.Fatal(err)
				}
			case common.MODE_CLOUD:
				if err := gorm.NewGormCloudInitializer(App.Logger, db, App.Location).Initialize(); err != nil {
					log.Fatal(err)
				}
			default:
				App.Logger.Fatalf("Unsupported mode: %s", App.Mode)
			}

			log.Println("Database initialized")
			return
		}

		if ShowConfigFlag {

			if DatastoreType == "memory" || DatastoreType == "sqlite" || DatastoreType == "postgres" || DatastoreType == "cockroach" {
				App.InitLogFile(os.Getuid(), os.Getgid())
				App.InitGormDB()
				serverConfig, _, _, _, _, err := builder.NewGormConfigBuilder(App).Build()
				if err != nil {
					App.Logger.Fatal(err)
				}
				//spew.Dump(serverConfig)
				var data []byte
				if ConfigOutputFormat == "yaml" {
					data, err = yaml.Marshal(serverConfig)
				} else if ConfigOutputFormat == "json" {
					data, err = json.Marshal(serverConfig)
				} else {
					App.Logger.Fatal("Unsupported output format: %s", ConfigOutputFormat)
				}
				if err != nil {
					App.Logger.Fatal(err)
				}
				fmt.Printf("%+v\n", string(data))
			} else {
				fmt.Printf("%+v\n", App.Config)
			}
		}

		if CompressConfigFlag {
			compressor := util.NewCompressor()
			var data []byte
			var err error
			if DatastoreType == "yaml" {
				data, err = yaml.Marshal(App.Config)
			} else if DatastoreType == "json" {
				data, err = json.Marshal(App.Config)
			} else {
				App.Logger.Fatalf("Unsupported datastore type: %s", DatastoreType)
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
