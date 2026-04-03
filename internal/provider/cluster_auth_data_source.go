// Copyright (c) 2026
// Licensed under the Mozilla Public License v2.0

package provider

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oci_common "github.com/oracle/oci-go-sdk/v65/common"
)

const (
	clusterAuthBaseURL    = "https://containerengine.%s.oraclecloud.com/cluster_request/%s"
	accessTokenExpiration = 4 * time.Minute
)

type clusterAuthResult struct {
	Token      string
	Expiration string
}

func dataSourceClusterAuth() *schema.Resource {
	return &schema.Resource{
		Description: "Generates a short-lived authentication token for an Oracle Kubernetes Engine (OKE) cluster.",
		Read:        readClusterAuth,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The OCID of the OKE cluster.",
			},
			"refresh_trigger": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Arbitrary value that forces the data source to be re-evaluated when it changes. Set this to `timestamp()` to defer token generation until apply and get the freshest token Terraform can provide.",
			},
			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The generated base64 URL-safe authentication token.",
			},
			"expiration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The HTTP date timestamp when the generated token expires.",
			},
		},
	}
}

func readClusterAuth(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	if clusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}

	providerMeta := meta.(*providerMetadata)
	region, err := providerMeta.ConfigProvider.Region()
	if err != nil {
		return err
	}

	createdAt := time.Now().UTC()
	result, err := generateClusterAuthToken(providerMeta.ContainerEngineClient.Signer, region, clusterID, createdAt)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s,%s", clusterID, createdAt.Format(time.RFC3339Nano)))
	if err := d.Set("token", result.Token); err != nil {
		return err
	}
	return d.Set("expiration", result.Expiration)
}

func generateClusterAuthToken(signer oci_common.HTTPRequestSigner, region, clusterID string, createdAt time.Time) (*clusterAuthResult, error) {
	requestURL := fmt.Sprintf(clusterAuthBaseURL, region, clusterID)

	signedRequest, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	signedRequest.Header.Set("date", createdAt.Format(http.TimeFormat))

	if err := signer.Sign(signedRequest); err != nil {
		return nil, err
	}

	tokenRequest, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	query := tokenRequest.URL.Query()
	query.Set("authorization", signedRequest.Header.Get("authorization"))
	query.Set("date", signedRequest.Header.Get("date"))
	tokenRequest.URL.RawQuery = query.Encode()

	return &clusterAuthResult{
		Token:      base64.URLEncoding.EncodeToString([]byte(tokenRequest.URL.String())),
		Expiration: createdAt.Add(accessTokenExpiration).Format(http.TimeFormat),
	}, nil
}
