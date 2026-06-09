package service

import "github.com/scality/nodeip-discovery/pkg/domain"

type InterfaceLister interface {
	// ListInterfaces returns a list of network interfaces
	ListInterfaces() ([]domain.IfaceAddrs, error)
}
