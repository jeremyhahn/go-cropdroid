//go:build !cluster
// +build !cluster

package cmd

import (
	"github.com/jeremyhahn/go-cropdroid/builder"
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/webservice"
	"github.com/spf13/cobra"
)

func init() {
	// standaloneCmd.PersistentFlags().StringVarP(&DeviceStore, "device-store", "", "datastore", "Where to store metrics [ datastore | redis ]")
	// DeviceStore = strings.ToLower(DeviceStore)

	rootCmd.AddCommand(standaloneCmd)
}

var standaloneCmd = &cobra.Command{
	Use:   "standalone",
	Short: "Run CropDroid Server in standalone mode",
	Long: `Starts the cropdroid real-time protection and notification service
	in "standalone mode". In standalone mode, data can be stored in a highly available,
	fauilt-tolerant database but the cropdroid service itself will not be fault-tolerant.
	Evaluation versions (no license file) are restricted to a single farm, device and
	admin user account without database persistence. This means your data will be lost and
	the system will return to default settings following a reboot. Licensed versions provide
	support for SQLite, MySQL, PostgreSQL, and CockroachDB (with support for experimental
	changefeeds to enable real-time notifications).`,
	Run: func(cmd *cobra.Command, args []string) {

		App.Mode = Mode

		rsaKeyPair, serviceRegistry, restServices, farmTickerProvisionerChan, err := builder.NewGormConfigBuilder(
			App, DeviceDataStore, AppStateTTL, AppStateTick, DatabaseInit).Build()

		if err != nil {
			App.Logger.Fatal(err)
		}

		App.KeyPair = rsaKeyPair

		farmServices := serviceRegistry.GetFarmServices()

		for _, farmService := range farmServices {
			go farmService.Run()
		}

		webserver := webservice.NewWebserver(App, serviceRegistry, restServices, farmTickerProvisionerChan)
		go webserver.Run()
		go webserver.RunProvisionerConsumer()

		if changefeedService := serviceRegistry.GetChangefeedService(); changefeedService != nil {
			changefeedService.Subscribe()
		}

		serviceRegistry.GetEventLogService(0).Create(0, common.CONTROLLER_TYPE_SERVER, "System", "Startup")

		done := make(chan error, 1)
		<-done
	},
}
