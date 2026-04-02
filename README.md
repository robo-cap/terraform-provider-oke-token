# terraform-provider-oke-token

Minimal Terraform provider for generating OKE cluster authentication tokens.

This repository intentionally contains only:

- Provider authentication configuration compatible with the OCI provider auth modes
- A single data source: `oketoken_cluster_auth`

It intentionally does not contain the upstream OCI provider's:

- Resources and unrelated data sources
- Framework provider and mux server
- Export/discovery commands
- Retry/tag/export configuration
- Test-only and service-specific client registries

## Supported authentication methods

- `ApiKey`
- `InstancePrincipal`
- `InstancePrincipalWithCerts`
- `SecurityToken`
- `ResourcePrincipal`
- `OKEWorkloadIdentity`

## Example

```hcl
terraform {
  required_providers {
    oketoken = {
      source  = "robo-cap/oke-token"
      version = "0.1.0"
    }
  }
}

provider "oketoken" {}

data "oketoken_cluster_auth" "cluster" {
  cluster_id = "ocid1.cluster.oc1..example"
}

output "token" {
  value     = data.oketoken_cluster_auth.cluster.token
  sensitive = true
}
```

## Local build

```bash
./build.sh
```

## Documentation generation

This repository follows the Terraform provider documentation layout used in the HashiCorp HashiCups tutorial:

- `tools/` contains the `go generate` entry point for `tfplugindocs`
- `examples/provider/` contains the provider configuration example
- `examples/data-sources/` contains data source examples used by generated docs
- `docs/` contains the generated provider and data source documentation

Generate or refresh the docs with:

```bash
go generate ./tools
```
