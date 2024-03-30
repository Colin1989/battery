package cluster

import (
	"fmt"
	"log/slog"

	"github.com/colin1989/battery/blog"

	"github.com/colin1989/battery/actor"
)

type Cluster struct {
	ActorSystem   *actor.ActorSystem
	Config        *Config
	Discovery     ServiceDiscovery
	ServersByType map[string]map[string]*Server
}

func New(as *actor.ActorSystem,
	config *Config,
	discovery ServiceDiscovery) *Cluster {
	c := &Cluster{}
	c.ActorSystem = as
	c.Config = config
	c.Discovery = discovery
	c.ActorSystem.ProcessRegistry.Address = fmt.Sprintf("%v:%v", c.Config.Address, c.Config.Port)

	return c
}

func (c *Cluster) UpdateServers() {

}

func (c *Cluster) Shutdown() {
	err := c.Discovery.Shutdown()
	if err != nil {
		c.Logger().Error("discovery shutdown error ", blog.ErrAttr(err))
	}
}

func (c *Cluster) Logger() *slog.Logger {
	return c.ActorSystem.Logger()
}
