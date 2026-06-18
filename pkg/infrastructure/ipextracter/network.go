package ipextracter

import (
	"log/slog"
	"maps"
	"net"
	"net/netip"
	"slices"
	"strings"

	"github.com/scality/nodeip-discovery/pkg/domain"
	"github.com/scality/nodeip-discovery/pkg/service"
)

type Network struct {
	logger     *slog.Logger
	cidrs      []netip.Prefix
	excludeIPs []netip.Addr
}

func NewNetwork(
	logger *slog.Logger,
	cidrs []netip.Prefix,
	excludeIPs []netip.Addr,
) *Network {
	l := logger.With(
		slog.String("ipdiscoverer", "localnetwork"),
	)
	return &Network{
		logger:     l,
		cidrs:      cidrs,
		excludeIPs: excludeIPs,
	}
}

var _ service.IPExtracter = &Network{}

// ExtractIPs returns the sorted, deduplicated list of local IPs assigned
// to up, non-loopback interfaces that fall into any of the configured CIDRs.
// Link-local and IPv6 addresses are excluded.
func (h *Network) ExtractIPs(ifaces []domain.IfaceAddrs) string {
	seen := make(map[netip.Addr]struct{})
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		for _, raw := range iface.Addresses {
			ipnet, ok := raw.(*net.IPNet)
			if !ok {
				continue
			}
			addr, ok := netip.AddrFromSlice(ipnet.IP)
			if !ok {
				continue
			}
			addr = addr.Unmap()
			if _, found := seen[addr]; found {
				continue
			}
			if !addr.Is4() {
				continue
			}
			if addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() {
				continue
			}
			if slices.Contains(h.excludeIPs, addr) {
				continue
			}
			for _, c := range h.cidrs {
				if c.Contains(addr) {
					seen[addr] = struct{}{}
					break
				}
			}
		}
	}
	out := slices.SortedFunc(maps.Keys(seen), func(a, b netip.Addr) int { return a.Compare(b) })

	// When no IP addresses are found, log it and return an empty string.
	if len(out) == 0 {
		h.logger.Info("no IP addresses found")
	}
	return formatAddrs(out)
}

func formatAddrs(addrs []netip.Addr) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = a.String()
	}
	return strings.Join(parts, ",")
}
