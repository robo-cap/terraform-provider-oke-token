// Copyright (c) 2026
// Licensed under the Mozilla Public License v2.0

package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/robo-cap/terraform-provider-oke-token/internal/provider"
)

func main() {
	if err := tf5server.Serve(provider.ProviderAddress, provider.New().GRPCProvider); err != nil {
		log.Fatal(err)
	}
}
