terraform {
  required_providers {
    oketoken = {
      source = "robo-cap/oke-token"
    }
  }
}

provider "oketoken" {
  auth             = "ApiKey"
  tenancy_ocid     = "ocid1.tenancy.oc1..exampleuniqueID"
  user_ocid        = "ocid1.user.oc1..exampleuniqueID"
  fingerprint      = "20:3b:97:13:55:1c:aa:bb:cc:dd:ee:ff:00:11:22:33"
  private_key_path = "~/.oci/oci_api_key.pem"
  region           = "us-ashburn-1"
}
