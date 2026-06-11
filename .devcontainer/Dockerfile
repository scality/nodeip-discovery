FROM mcr.microsoft.com/devcontainers/go:1.26-trixie@sha256:c249440b024e82a569726ab41e60400237665d1f259fb575fd707b04b1b44e55

USER vscode

# renovate: datasource=github-releases depName=golangci/golangci-lint extractVersion=^v(?<version>.*)$
ARG GOLANGCI_LINT_VERSION=2.11.4

RUN curl -sSfL  https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sudo sh -s -- v${GOLANGCI_LINT_VERSION}
