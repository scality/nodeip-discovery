package di

import (
	"github.com/scality/nodeip-discovery/pkg/infrastructure/interfacelister"
	"github.com/scality/nodeip-discovery/pkg/service"
)

func (c *Container) getInterfaceLister() service.InterfaceLister {
	if c.interfaceLister == nil {
		c.interfaceLister = interfacelister.NewLocal(
			c.GetLogger(),
		)
	}
	return c.interfaceLister
}
