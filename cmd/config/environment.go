package config

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/scality/go-errors"
	"github.com/sethvargo/go-envconfig"
)

// ApplicationVersion is the version of the application.
// It is set at build time using ldflags.
//
//nolint:gochecknoglobals // This is a constant.
var ApplicationVersion = "dev"

const (
	ApplicationName = "nodeip-discovery"
)

var ErrConfigurationLoading error = errors.New("Configuration Loading Error")

type (
	Environment struct {
		Logger LoggerConfig `env:",prefix=LOGGER_"`

		NodeName        string `env:"NODE_NAME"`
		ApplicationName string

		AnnotationRootKey string        `env:"ANNOTATION_ROOT_KEY, default=loadbalancer.scality.com"`
		ResyncInterval    time.Duration `env:"RESYNC_INTERVAL, default=30s"`
		MaxAttempts       int           `env:"MAX_ATTEMPTS, default=10"`

		CIDRs      []netip.Prefix
		ExcludeIPs []netip.Addr
	}

	// LoggerConfig holds the logging configuration loaded from the environment.
	LoggerConfig struct {
		LogLevel string `env:"LOG_LEVEL, default=info"`
	}
)

func NewEnvironment(ctx context.Context, cidrList string, excludeIPList string) (*Environment, error) {
	cidrs, err := validateCIDRs(cidrList)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	excludeIPs, err := validateExcludeIPs(excludeIPList)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	cfg := &Environment{
		CIDRs:           cidrs,
		ExcludeIPs:      excludeIPs,
		ApplicationName: ApplicationName,
	}

	err = cfg.Load(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return cfg, nil
}

func (cfg *Environment) Load(ctx context.Context) error {
	err := envconfig.Process(ctx, cfg)
	if err != nil {
		return errors.Wrap(ErrConfigurationLoading,
			errors.WithDetail("failed to process environment variables"),
			errors.CausedBy(err),
		)
	}

	if cfg.ResyncInterval <= 0 {
		return errors.Wrap(ErrConfigurationLoading,
			errors.WithDetail("RESYNC_INTERVAL must be greater than zero"),
		)
	}

	if cfg.MaxAttempts <= 0 {
		return errors.Wrap(ErrConfigurationLoading,
			errors.WithDetail("MAX_ATTEMPTS must be greater than zero"),
		)
	}

	if cfg.NodeName == "" {
		return errors.Wrap(ErrConfigurationLoading,
			errors.WithDetail("NODE_NAME environment variable is required"),
		)
	}

	return nil
}

func validateCIDRs(cidrList string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, raw := range strings.Split(cidrList, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		p, err := netip.ParsePrefix(raw)
		if err != nil {
			return nil, errors.Wrap(ErrConfigurationLoading,
				errors.WithDetail("invalid CIDR"),
				errors.WithProperty("cidr", raw),
				errors.CausedBy(err),
			)
		}
		if !p.Addr().Is4() {
			return nil, errors.Wrap(ErrConfigurationLoading,
				errors.WithDetail("only IPv4 CIDRs are supported"),
				errors.WithProperty("cidr", raw),
			)
		}
		prefixes = append(prefixes, p.Masked())
	}
	if len(prefixes) == 0 {
		return nil, errors.Wrap(ErrConfigurationLoading,
			errors.WithDetail("at least one valid CIDR must be provided"),
		)
	}

	return prefixes, nil
}

func validateExcludeIPs(excludeIPs string) ([]netip.Addr, error) {
	var addrs []netip.Addr
	for _, raw := range strings.Split(excludeIPs, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		addr, err := netip.ParseAddr(raw)
		if err != nil {
			return nil, errors.Wrap(ErrConfigurationLoading,
				errors.WithDetail("invalid exclude IP"),
				errors.WithProperty("exclude-ip", raw),
				errors.CausedBy(err),
			)
		}
		if !addr.Is4() {
			return nil, errors.Wrap(ErrConfigurationLoading,
				errors.WithDetail("only IPv4 IPs are supported"),
				errors.WithProperty("exclude-ip", raw),
			)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
