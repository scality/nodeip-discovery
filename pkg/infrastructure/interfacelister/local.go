package interfacelister

import (
	"log/slog"
	"net"

	"github.com/scality/go-errors"
	"github.com/scality/nodeip-discovery/pkg/domain"
	"github.com/scality/nodeip-discovery/pkg/service"
)

type Local struct {
	logger *slog.Logger
}

func NewLocal(logger *slog.Logger) *Local {
	return &Local{logger: logger}
}

var _ service.InterfaceLister = &Local{}

// systemInterfaces is the production interfaceLister: it enumerates the host's
// interfaces via the net package. Interfaces whose addresses cannot be listed
// are logged and skipped rather than failing the whole discovery.
func (l *Local) ListInterfaces() ([]domain.IfaceAddrs, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, errors.Wrap(domain.ErrAddressesDiscoveryFailed,
			errors.WithDetail("failed to list interfaces"),
			errors.CausedBy(err),
		)
	}

	out := make([]domain.IfaceAddrs, 0, len(ifaces))
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			l.logger.Warn("failed to list interface addresses",
				slog.String("interface", iface.Name),
				slog.Any("error", err),
			)
			continue
		}
		out = append(out, domain.IfaceAddrs{
			Name:      iface.Name,
			Flags:     iface.Flags,
			Addresses: addrs,
		})
	}
	return out, nil
}
