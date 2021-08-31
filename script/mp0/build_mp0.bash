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
go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/mp0-c" "${PROJECT_ROOT}/cli/mp0/client"
go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/mp0-s" "${PROJECT_ROOT}/cli/mp0/server"

