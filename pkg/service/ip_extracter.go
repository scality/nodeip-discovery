package service

import "github.com/scality/nodeip-discovery/pkg/domain"

type IPExtracter interface {
	// ExtractIPs returns a comma-separated string of the IP addresses from a given list of interfaces
	// matching a configured list of CIDRs
	ExtractIPs(ifaces []domain.IfaceAddrs) string
}
