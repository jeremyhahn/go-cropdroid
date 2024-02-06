package test

import (
	"os"
	"sync"
	"time"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/common"
	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

var CurrentTest *DeviceTest = &DeviceTest{mutex: &sync.Mutex{}}

type DeviceTest struct {
	mutex    *sync.Mutex
	db       gormstore.GormDB
	gorm     *gorm.DB
	logger   *logging.Logger
	location *time.Location
}

func NewUnitTestSession() *app.App {
	wd, _ := os.Getwd()

	logger := logging.MustGetLogger(common.APPNAME)
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logging.SetLevel(logging.DEBUG, "")

	location, _ := time.LoadLocation("America/New_York")

	return &app.App{
		HomeDir:  wd,
		Logger:   logger,
		Location: location}
}
