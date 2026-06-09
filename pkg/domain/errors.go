package domain

import "github.com/scality/go-errors"

var (
	ErrReconciliationFailed     error = errors.New("Reconciliation Failed")
	ErrAddressesDiscoveryFailed error = errors.New("Addresses Discovery Failed")
)
