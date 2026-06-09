package usecase

import (
	"context"
	"log/slog"

	"github.com/scality/go-errors"
	"github.com/scality/nodeip-discovery/pkg/service"
)

type DiscoverNodeIP struct {
	logger          *slog.Logger
	ipExtracter     service.IPExtracter
	nodeAnnotater   service.NodeAnnotater
	interfaceLister service.InterfaceLister
}

func NewDiscoverNodeIP(
	logger *slog.Logger,
	ipExtracter service.IPExtracter,
	nodeAnnotater service.NodeAnnotater,
	interfaceLister service.InterfaceLister,
) *DiscoverNodeIP {
	l := logger.With(
		slog.String("usecase", "discover_nodeip"),
	)
	return &DiscoverNodeIP{
		logger:          l,
		ipExtracter:     ipExtracter,
		nodeAnnotater:   nodeAnnotater,
		interfaceLister: interfaceLister,
	}
}

func (uc *DiscoverNodeIP) Execute(ctx context.Context) error {
	uc.logger.Info("Discovering node IP")

	ifaces, err := uc.interfaceLister.ListInterfaces()
	if err != nil {
		return errors.Wrap(err,
			errors.WithDetail("failed to list interfaces"),
		)
	}

	ips := uc.ipExtracter.ExtractIPs(ifaces)

	err = uc.nodeAnnotater.AnnotateNode(ctx, ips)
	if err != nil {
		return errors.Wrap(err,
			errors.WithDetail("failed to annotate node"),
		)
	}

	uc.logger.Info("Node IP discovered successfully")
	return nil
}
