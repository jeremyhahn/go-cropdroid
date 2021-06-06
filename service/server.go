package service

import "github.com/jeremyhahn/go-cropdroid/app"

type ServerService interface {
}

type DefaultServerService struct {
	app *app.App
}
