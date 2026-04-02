terraform {
  required_providers {
    oke-token = {
      source  = "robo-cap/oke-token"
      version = "1.0.0"
    }
  }
}

provider "oke-token" {
  region = "eu-frankfurt-1"
}

data "oketoken_cluster_auth" "cluster" {
  cluster_id = "ocid1.cluster.oc1.eu-frankfurt-1.aaaaaaaapwojsv7pqypfq4uf2pkqyhuyyouhtu556fyomhnoicg243n6z6aa"
}

output "token" {
  value     = data.oketoken_cluster_auth.cluster.token
  sensitive = true
}

output "expiration" {
  value = data.oketoken_cluster_auth.cluster.expiration
}
