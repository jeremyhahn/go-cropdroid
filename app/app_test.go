package app

import (
	"os"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	logging "github.com/op/go-logging"
)

func NewUnitTest() *App {

	logger := logging.MustGetLogger(common.APPNAME)
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(stdout)
	logging.SetLevel(logging.DEBUG, "")

	location, _ := time.LoadLocation("America/New_York")

	return &App{
		Logger:   logger,
		Location: location}
}
