package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/scality/nodeip-discovery/cmd/config"
	"github.com/scality/nodeip-discovery/pkg/infrastructure/di"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var cidrList string
	var kubeconfig string
	var excludeIPList string
	flag.StringVar(&cidrList, "cidrs", "", "Comma-separated list of CIDRs to match against local interfaces (IPv4 only)")
	flag.StringVar(&excludeIPList, "exclude-ips", "", "Comma-separated list of IPs to exclude from the discovery (IPv4 only)")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file. If empty, in-cluster config is used")
	flag.Parse()
	if strings.TrimSpace(cidrList) == "" {
		fmt.Fprintf(os.Stderr, "cidrs is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load configuration from environment variables.
	cfg, err := config.NewEnvironment(ctx, cidrList, excludeIPList)
	if err != nil {
		// %s renders go-errors' human-readable form; the default %v
		// (what Fprintln would use) emits JSON, which is for logs.
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	// Initiate a Kubernetes client to annotate the Node
	var restCfg *rest.Config
	if kubeconfig != "" {
		restCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to build config from kubeconfig %q: %s\n", kubeconfig, err)
			os.Exit(1)
		}
	} else {
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get in-cluster config: %s\n", err)
			os.Exit(1)
		}
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build kube client: %s\n", err)
		os.Exit(1)
	}

	// Initialize the discovery service
	container := di.NewContainer(cfg, client)
	logger := container.GetLogger()

	// Run the discovery use case periodically
	ticker := time.NewTicker(cfg.ResyncInterval)
	defer ticker.Stop()

	// Retry the discovery use case up to maxAttempts times
	attempts := 1
	tick := func() {
		if err := container.GetDiscoverNodeIPUseCase().Execute(ctx); err != nil {
			logger.Error("discover node IP failed", "error", err, "attempts", attempts)
			attempts++
			return
		}
		// Reset the attempts counter after successful discovery
		attempts = 1
	}
	tick()
	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down")
			return
		case <-ticker.C:
			tick()
			if attempts > cfg.MaxAttempts {
				logger.Error("max attempts reached, exiting")
				os.Exit(1)
			}
		}
	}
}
