package domain

import "net"

// IfaceAddrs bundles a network interface's flags with its assigned addresses.
// It is the unit consumed by IPExtracter and is produced by an InterfaceLister.
type IfaceAddrs struct {
	Name      string
	Flags     net.Flags
	Addresses []net.Addr
}
