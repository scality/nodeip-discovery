package di

import (
	"github.com/scality/nodeip-discovery/pkg/infrastructure/ipextracter"
	"github.com/scality/nodeip-discovery/pkg/service"
)

func (c *Container) getIPDiscoverer() service.IPExtracter {
	if c.ipExtracter == nil {
		c.ipExtracter = ipextracter.NewNetwork(
			c.GetLogger(),
			c.config.CIDRs,
			c.config.ExcludeIPs,
		)
	}
	return c.ipExtracter
}
