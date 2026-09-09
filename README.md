# nodeip-discovery

[![Post Merge](https://github.com/scality/nodeip-discovery/actions/workflows/post-merge.yaml/badge.svg)](https://github.com/scality/nodeip-discovery/actions/workflows/post-merge.yaml)
[![GitHub release](https://img.shields.io/github/v/release/scality/nodeip-discovery)](https://github.com/scality/nodeip-discovery/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/scality/nodeip-discovery)](go.mod)
[![License](https://img.shields.io/github/license/scality/nodeip-discovery)](LICENSE)

A Kubernetes node-level agent that discovers the local network interfaces of the
node it runs on, filters their IPv4 addresses against a configured set of CIDRs
and a given list of excluded IPs and writes the result to a node annotation.

It is designed to run as a DaemonSet so that every node advertises the set of
addresses it owns within one or more known networks (for example, a load-balancer
data network). Other controllers can then read the annotation to route traffic to
the right nodes.

## How it works

On a fixed interval the agent:

1. Lists the node's network interfaces.
2. Keeps the IPv4 addresses of up, non-loopback interfaces that fall within any of
   the configured CIDRs (link-local and IPv6 addresses, and any address listed in
   `-exclude-ips`, are excluded).
3. Patches the node with the sorted, deduplicated, comma-separated list under the
   configured annotation key. When nothing matches, the annotation is set to an
   empty value rather than failing.

For the layered architecture and the rationale behind these behaviours, see
[DESIGN.md](DESIGN.md).

## Usage

The binary is published as a distroless, non-root image and expects to run inside
the cluster (it uses the in-cluster Kubernetes config by default). The node name
is provided through the downward API.

```sh
# In-cluster (typical DaemonSet container args)
nodeip-discovery -cidrs=10.99.0.0/24,192.168.0.0/16

# Exclude specific addresses from the matched set
nodeip-discovery -cidrs=10.99.0.0/24 -exclude-ips=10.99.0.1,10.99.0.2

# Out-of-cluster, against a specific kubeconfig
nodeip-discovery -cidrs=10.99.0.0/24 -kubeconfig=$HOME/.kube/config
```

The container needs RBAC permissions to `get` and `patch` the `Node` object it
runs on, and `NODE_NAME` populated from `spec.nodeName` via the downward API.

### Configuration

| Flag / Env var       | Type            | Default                   | Description                                                                 |
| -------------------- | --------------- | ------------------------- | --------------------------------------------------------------------------- |
| `-cidrs`             | flag (required) | —                         | Comma-separated list of IPv4 CIDRs to match against local interfaces.       |
| `-exclude-ips`       | flag            | _(none)_                  | Comma-separated list of IPv4 addresses to exclude from discovery.           |
| `-kubeconfig`        | flag            | _(in-cluster)_            | Path to a kubeconfig file. When empty, the in-cluster config is used.       |
| `NODE_NAME`          | env (required)  | —                         | Name of the node to annotate (set via the downward API).                    |
| `ANNOTATION_ROOT_KEY`| env             | `loadbalancer.scality.com`| Annotation key written on the node.                                         |
| `RESYNC_INTERVAL`    | env             | `30s`                     | How often discovery runs. Must be greater than zero.                        |
| `MAX_ATTEMPTS`       | env             | `10`                      | Consecutive failures tolerated before the process exits non-zero.           |
| `LOGGER_LOG_LEVEL`   | env             | `info`                    | slog level: `debug`, `info`, `warn`, `error`.                               |

## Building

```sh
make docker-build IMG=nodeip-discovery:latest
```

See the [Makefile](Makefile) for the full list of targets

## Contributing

Contributions are welcome — please read [CONTRIBUTING.md](CONTRIBUTING.md) for
the development setup, testing, and pull-request process.

## Documentation

- [DESIGN.md](DESIGN.md) — architecture and internals.
- [CONTRIBUTING.md](CONTRIBUTING.md) — development workflow and conventions.
