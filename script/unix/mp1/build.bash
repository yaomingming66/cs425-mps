#!/usr/bin/env bash

set -o pipefail
set -o errexit
set -o xtrace

if ! [ -x "$(command -v git)" ]; then
    printf "%s\n" 'Error: git is not installed.' >&2
    exit 1
fi

if ! [ -x "$(command -v go)" ]; then
    printf "%s\n" 'Error: go is not installed.' >&2
    exit 1
fi

# GOOS=linux GOARCH="amd64"
PROJECT_ROOT=$(git rev-parse --show-toplevel)

GOOS="linux" GOARCH="amd64" go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/mp1-linux-amd64" "${PROJECT_ROOT}/cli/mp1/node"
GOOS="darwin" GOARCH="arm64" go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/mp1-darwin-arm64" "${PROJECT_ROOT}/cli/mp1/node"
GOOS="windows" GOARCH="amd64" go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/mp1-windows-amd64.exe" "${PROJECT_ROOT}/cli/mp1/node"