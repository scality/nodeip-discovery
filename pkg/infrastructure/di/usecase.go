package di

import "github.com/scality/nodeip-discovery/pkg/usecase"

func (c *Container) GetDiscoverNodeIPUseCase() *usecase.DiscoverNodeIP {
	if c.discoverNodeIPUseCase == nil {
		c.discoverNodeIPUseCase = usecase.NewDiscoverNodeIP(
			c.GetLogger(),
			c.getIPDiscoverer(),
			c.getNodeAnnotater(),
			c.getInterfaceLister(),
		)
	}
	return c.discoverNodeIPUseCase
}
