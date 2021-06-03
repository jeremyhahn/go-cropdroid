package builder

import (
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/jeremyhahn/cropdroid/state"
	"github.com/jeremyhahn/cropdroid/webservice/rest"
)

type ConfigBuilder interface {
	Build() (config.ServerConfig, service.ServiceRegistry, []rest.RestService, state.ControllerIndex, state.ChannelIndex, error)
}
