// +build !cluster

package cmd

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(standaloneCmd)
}

var standaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "Run CropDroid Server in standalone mode",
	Long: `Starts the cropdroid real-time protection and notification service
	in "standalone mode". In standalone mode, data can be stored in a highly available,
	fauilt-tolerant database but the cropdroid service itself will not be fault-tolerant.
	Evaluation versions (no license file) are restricted to a single farm, controller and
	admin user account without database persistence. This means your data will be lost and
	the system will return to default settings following a reboot. Licensed versions provide
	support for SQLite, MySQL, PostgreSQL, and CockroachDB (with support for experimental
	changefeeds to enable real-time notifications).`,
	Run: func(cmd *cobra.Command, args []string) {

		App.Mode = common.MODE_STANDALONE
		App.InitGormDB()
		App.FarmStore = state.NewMemoryFarmStore(App.Logger, 1, AppStateTTL, time.Duration(AppStateTick))
		App.ConfigStore = state.NewMemoryConfigStore(1)

		if MetricDatastore == "datastore" {
			App.MetricDatastore = gorm.NewControllerStateDAO(App.Logger, App.GORM, App.GORMInitParams.Engine, App.Location)
		} else if MetricDatastore == "redis" {
			App.MetricDatastore = datastore.NewRedisControllerStateDAO(":6379", "")
		}

		serverConfig, serviceRegistry, restServices, controllerIndex, channelIndex, err := builder.NewGormConfigBuilder(App).Build()
		if err != nil {
			App.Logger.Fatal(err)
		}

		App.Config = serverConfig.(*config.Server)
		App.ControllerIndex = controllerIndex
		App.ChannelIndex = channelIndex

		farmServices := serviceRegistry.GetFarmServices()
		if len(farmServices) != 1 {
			App.Logger.Fatal("Invalid standalone farm configuration")
		}

		for _, farmService := range farmServices {
			go farmService.Run()
		}

		webserver := webservice.NewWebserver(App, serviceRegistry, restServices)
		go webserver.Run()

		if changefeedService := serviceRegistry.GetChangefeedService(); changefeedService != nil {
			changefeedService.Subscribe()
		}

		done := make(chan error, 1)
		<-done
	},
}
