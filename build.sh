#!/bin/bash

set -euo pipefail

export GOARCH="amd64"
export GOOS="linux"
export CGO_ENABLED="0"

go build -v -ldflags='-s' -o bin/terraform-provider-oke-token_v1.0.0

mkdir -p "$HOME/.terraform.d/plugins/registry.terraform.io/robo-cap/oke-token/1.0.0/linux_amd64/"
ln -sf "$(pwd)/bin/terraform-provider-oke-token_v1.0.0" \
  "$HOME/.terraform.d/plugins/registry.terraform.io/robo-cap/oke-token/1.0.0/linux_amd64/terraform-provider-oke-token_v1.0.0"
