#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$ROOT_DIR/.tfplugindocs-work"
PLUGIN_DIR="$WORK_DIR/plugins/registry.terraform.io/robo-cap/oke-token/0.0.1/linux_amd64"
SCHEMA_FILE="$WORK_DIR/providers-schema.json"

cleanup() {
  rm -rf "$WORK_DIR"
}

trap cleanup EXIT

mkdir -p "$PLUGIN_DIR"

go build -o "$PLUGIN_DIR/terraform-provider-oke-token" "$ROOT_DIR"

cat > "$WORK_DIR/provider.tf" <<'EOF'
terraform {
  required_providers {
    oketoken = {
      source = "robo-cap/oke-token"
    }
  }
}

provider "oketoken" {}
EOF

terraform -chdir="$WORK_DIR" init -backend=false -get=false -plugin-dir=./plugins >/dev/null
terraform -chdir="$WORK_DIR" providers schema -json > "$SCHEMA_FILE"

cd "$ROOT_DIR"
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.21.0 generate --provider-name oke-token --providers-schema "$SCHEMA_FILE"
