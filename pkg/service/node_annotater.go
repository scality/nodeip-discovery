package service

import (
	"context"
)

type NodeAnnotater interface {
	// AnnotateNode annotates the node with the given list of IPs
	AnnotateNode(ctx context.Context, formattedIPs string) error
}
