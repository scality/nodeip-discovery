package nodeannotater

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/scality/go-errors"
	"github.com/scality/nodeip-discovery/pkg/domain"
	"github.com/scality/nodeip-discovery/pkg/service"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type Kubernetes struct {
	logger            *slog.Logger
	client            kubernetes.Interface
	nodeName          string
	annotationRootKey string
	fieldManager      string
}

func NewKubernetes(
	logger *slog.Logger,
	client kubernetes.Interface,
	nodeName string,
	annotationRootKey string,
	fieldManager string,
) *Kubernetes {
	return &Kubernetes{
		logger:            logger,
		client:            client,
		nodeName:          nodeName,
		annotationRootKey: annotationRootKey,
		fieldManager:      fieldManager,
	}
}

var _ service.NodeAnnotater = &Kubernetes{}

func (k *Kubernetes) AnnotateNode(ctx context.Context, ips string) error {
	node, err := k.client.CoreV1().Nodes().Get(ctx, k.nodeName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Wrap(domain.ErrReconciliationFailed,
				errors.WithDetail("node not found"),
			)
		}
		return errors.Wrap(domain.ErrReconciliationFailed,
			errors.WithDetail("failed to get node"),
			errors.CausedBy(err),
		)
	}
	if current, ok := node.Annotations[k.annotationRootKey]; ok && current == ips {
		k.logger.Debug("annotation already up-to-date", "annotation", k.annotationRootKey, "value", ips)
		return nil
	}

	patch, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]string{
				k.annotationRootKey: ips,
			},
		},
	})
	if err != nil {
		return errors.Wrap(domain.ErrReconciliationFailed,
			errors.WithDetail("failed to marshal patch"),
			errors.CausedBy(err),
		)
	}
	_, err = k.client.CoreV1().Nodes().Patch(
		ctx, k.nodeName, types.StrategicMergePatchType, patch,
		metav1.PatchOptions{FieldManager: k.fieldManager},
	)
	if err != nil {
		return errors.Wrap(domain.ErrReconciliationFailed,
			errors.WithDetail("failed to patch node"),
			errors.CausedBy(err),
		)
	}
	k.logger.Debug("updated annotation", "annotation", k.annotationRootKey, "value", ips)
	return nil
}
