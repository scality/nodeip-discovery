package di

import (
	"github.com/scality/nodeip-discovery/pkg/infrastructure/nodeannotater"
	"github.com/scality/nodeip-discovery/pkg/service"
)

func (c *Container) getNodeAnnotater() service.NodeAnnotater {
	if c.nodeAnnotater == nil {
		c.nodeAnnotater = nodeannotater.NewKubernetes(
			c.GetLogger(),
			c.client,
			c.config.NodeName,
			c.config.AnnotationRootKey,
			c.config.ApplicationName,
		)
	}
	return c.nodeAnnotater
}
