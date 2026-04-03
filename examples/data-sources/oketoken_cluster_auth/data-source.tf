terraform {
  required_providers {
    oketoken = {
      source = "robo-cap/oke-token"
    }
  }
}

provider "oketoken" {}

data "oketoken_cluster_auth" "example" {
  cluster_id      = "ocid1.cluster.oc1.iad.exampleuniqueID"
  refresh_trigger = timestamp()
}
