# nodeip-discovery

This is a **Go service that discovers all network interfaces of a Node and annotates that Node** (Kubernetes). It is packaged as a single distroless, non-root binary and follows a Kubernetes operator/controller-style layout. It contains:

- Entry point and command wiring (`cmd/`)
- Build-time configuration such as application name and version (`cmd/config/`, set via `-ldflags`)
- Containerized build via multi-stage `Dockerfile` (golang builder → `distroless/static:nonroot` runtime)
- CI workflows for build, test, e2e, release, promote, SBOM, and pre/post-merge checks (`.github/workflows/`)
- `Makefile` targets for `lint` (golangci-lint), `fmt`, `vet`, `test`, and docker build/push

Tech stack: Go 1.26, modules (`go.mod`), `golangci-lint` for linting, Renovate for dependency updates. Expect mostly `.go` files, plus YAML workflows and the `Dockerfile`.
