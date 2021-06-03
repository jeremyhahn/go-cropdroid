package service

import "github.com/jeremyhahn/cropdroid/app"

type ServerService interface {
}

type DefaultServerService struct {
	app *app.App
}
