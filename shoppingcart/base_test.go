package shoppingcart

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

var CurrentTest *TestSuite = &TestSuite{}
var Location *time.Location
var TestSuiteName = "cropdroid_shoppingcart_test"
var EnableDefaultFarm = true

type TestSuite struct {
	app      *app.App
	logger   *logging.Logger
	location *time.Location
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {

	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logger := logging.MustGetLogger(TestSuiteName)

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal(err)
	}
	Location = location

	app := &app.App{
		KeyDir:   "../keys",
		Logger:   logger,
		Location: Location}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("/etc/cropdroid/"))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.cropdroid/"))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		app.Logger.Errorf("%s", err)
	}

	viper.Unmarshal(app)

	CurrentTest.app = app
	CurrentTest.logger = logger
	CurrentTest.location = Location
}

func teardown() {
	// if CurrentTest != nil {
	// 	CurrentTest.Cleanup()
	// }
}
