provider "oke-token" {}

data "oketoken_cluster_auth" "example" {
  cluster_id = "ocid1.cluster.oc1.iad.exampleuniqueID"
}
