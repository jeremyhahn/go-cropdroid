package builder

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/webservice/rest"
)

type ConfigBuilder interface {
	Build() (config.ServerConfig, service.ServiceRegistry, []rest.RestService, state.ControllerIndex, state.ChannelIndex, error)
}
