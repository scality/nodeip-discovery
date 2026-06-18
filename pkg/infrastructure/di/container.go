package di

import (
	"log/slog"

	"github.com/scality/nodeip-discovery/cmd/config"
	"github.com/scality/nodeip-discovery/pkg/service"
	"github.com/scality/nodeip-discovery/pkg/usecase"
	"k8s.io/client-go/kubernetes"
)

type Container struct {
	config *config.Environment

	logger *slog.Logger

	client                kubernetes.Interface
	ipExtracter           service.IPExtracter
	nodeAnnotater         service.NodeAnnotater
	interfaceLister       service.InterfaceLister
	discoverNodeIPUseCase *usecase.DiscoverNodeIP
}

func NewContainer(
	cfg *config.Environment,
	client kubernetes.Interface,
) *Container {
	return &Container{
		config: cfg,
		client: client,
	}
}
