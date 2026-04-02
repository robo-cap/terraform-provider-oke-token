// Copyright (c) 2026
// Licensed under the Mozilla Public License v2.0

package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	oci_common "github.com/oracle/oci-go-sdk/v65/common"
	oci_containerengine "github.com/oracle/oci-go-sdk/v65/containerengine"
)

const (
	Version                 = "0.1.0"
	ProviderTypeName        = "oke-token"
	ProviderSourceAddress   = "robo-cap/oke-token"
	ProviderAddress         = "registry.terraform.io/robo-cap/oke-token"
	PrimaryDataSourceName   = "oketoken_cluster_auth"
	AuthAPIKey              = "ApiKey"
	AuthInstancePrincipal   = "InstancePrincipal"
	AuthInstanceWithCerts   = "InstancePrincipalWithCerts"
	AuthSecurityToken       = "SecurityToken"
	AuthResourcePrincipal   = "ResourcePrincipal"
	AuthOKEWorkloadIdentity = "OKEWorkloadIdentity"
)

const (
	AuthAttrName               = "auth"
	TenancyOcidAttrName        = "tenancy_ocid"
	UserOcidAttrName           = "user_ocid"
	FingerprintAttrName        = "fingerprint"
	PrivateKeyAttrName         = "private_key"
	PrivateKeyPathAttrName     = "private_key_path"
	PrivateKeyPasswordAttrName = "private_key_password"
	RegionAttrName             = "region"
	ConfigFileProfileAttrName  = "config_file_profile"
)

type providerMetadata struct {
	ConfigProvider        oci_common.ConfigurationProvider
	ContainerEngineClient *oci_containerengine.ContainerEngineClient
}

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			AuthAttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Authentication mode. Supported values: `ApiKey`, `SecurityToken`, `InstancePrincipal`, `InstancePrincipalWithCerts`, `ResourcePrincipal`, and `OKEWorkloadIdentity`.",
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{tfVarName(AuthAttrName), ociVarName(AuthAttrName)}, AuthAPIKey),
				ValidateFunc: validation.StringInSlice([]string{AuthAPIKey, AuthSecurityToken, AuthInstancePrincipal, AuthInstanceWithCerts, AuthResourcePrincipal, AuthOKEWorkloadIdentity}, true),
			},
			TenancyOcidAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OCI tenancy OCID used by `ApiKey` and `SecurityToken` authentication.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(TenancyOcidAttrName), ociVarName(TenancyOcidAttrName)}, nil),
			},
			UserOcidAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OCI user OCID used by `ApiKey` and `SecurityToken` authentication.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(UserOcidAttrName), ociVarName(UserOcidAttrName)}, nil),
			},
			FingerprintAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Fingerprint for the API signing key used by `ApiKey` authentication.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(FingerprintAttrName), ociVarName(FingerprintAttrName)}, nil),
			},
			PrivateKeyAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "PEM-formatted API signing private key contents.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(PrivateKeyAttrName), ociVarName(PrivateKeyAttrName)}, nil),
			},
			PrivateKeyPathAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to the PEM-formatted API signing private key file.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(PrivateKeyPathAttrName), ociVarName(PrivateKeyPathAttrName)}, ""),
			},
			PrivateKeyPasswordAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password protecting the private key, when the key is encrypted.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(PrivateKeyPasswordAttrName), ociVarName(PrivateKeyPasswordAttrName)}, ""),
			},
			RegionAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OCI region, for example `us-ashburn-1`.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(RegionAttrName), ociVarName(RegionAttrName)}, nil),
			},
			ConfigFileProfileAttrName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Profile name in `~/.oci/config` when configuration is loaded from the OCI config file.",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{tfVarName(ConfigFileProfileAttrName), ociVarName(ConfigFileProfileAttrName)}, nil),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			PrimaryDataSourceName: dataSourceClusterAuth(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	configProvider, err := sdkConfigProviderFromResourceData(d)
	if err != nil {
		return nil, err
	}

	containerEngineClient, err := oci_containerengine.NewContainerEngineClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	configureBaseClient(&containerEngineClient.BaseClient, configProvider)

	return &providerMetadata{
		ConfigProvider:        configProvider,
		ContainerEngineClient: &containerEngineClient,
	}, nil
}

func configureBaseClient(client *oci_common.BaseClient, configProvider oci_common.ConfigurationProvider) {
	client.HTTPClient = buildHTTPClient()
	client.Signer = oci_common.DefaultRequestSigner(configProvider)
	client.UserAgent = fmt.Sprintf("%s/%s", ProviderTypeName, Version)
}

func tfVarName(attrName string) string {
	return "TF_VAR_" + attrName
}

func ociVarName(attrName string) string {
	return "OCI_" + strings.ToUpper(attrName)
}
