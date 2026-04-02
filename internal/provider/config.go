// Copyright (c) 2026
// Licensed under the Mozilla Public License v2.0

package provider

import (
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oci_common "github.com/oracle/oci-go-sdk/v65/common"
	oci_common_auth "github.com/oracle/oci-go-sdk/v65/common/auth"
)

const (
	defaultConfigDirName     = ".oci"
	defaultConfigFileName    = "config"
	boatTenancyOcidAttrName  = "boat_tenancy_ocid"
	testCertificatesLocation = "test_certificates_location"
)

func sdkConfigProviderFromResourceData(d *schema.ResourceData) (oci_common.ConfigurationProvider, error) {
	auth := strings.ToLower(d.Get(AuthAttrName).(string))
	profile := d.Get(ConfigFileProfileAttrName).(string)

	configProviders, err := authConfigProviders(d, auth)
	if err != nil {
		return nil, err
	}

	configProviders = append(configProviders, resourceDataConfigProvider{d})

	if profile == "" {
		configProviders = append(configProviders, oci_common.DefaultConfigProvider())
	} else {
		defaultPath := path.Join(homeFolder(), defaultConfigDirName, defaultConfigFileName)
		if err := checkProfile(profile, defaultPath); err != nil {
			return nil, err
		}
		configProviders = append(configProviders, oci_common.CustomProfileConfigProvider(defaultPath, profile))
	}

	return oci_common.ComposingConfigurationProvider(configProviders)
}

func authConfigProviders(d *schema.ResourceData, auth string) ([]oci_common.ConfigurationProvider, error) {
	switch auth {
	case strings.ToLower(AuthAPIKey):
		return nil, nil
	case strings.ToLower(AuthInstancePrincipal):
		region, ok := d.GetOkExists(RegionAttrName)
		if !ok || region.(string) == "" {
			return nil, fmt.Errorf("can not get %s from Terraform configuration (InstancePrincipal)", RegionAttrName)
		}

		return instancePrincipalProviders(region.(string))
	case strings.ToLower(AuthInstanceWithCerts):
		return instancePrincipalWithCertsProviders(d)
	case strings.ToLower(AuthSecurityToken):
		return securityTokenProviders(d)
	case strings.ToLower(AuthResourcePrincipal):
		return resourcePrincipalProviders(d)
	case strings.ToLower(AuthOKEWorkloadIdentity):
		cfg, err := oci_common_auth.OkeWorkloadIdentityConfigurationProvider()
		if err != nil {
			return nil, fmt.Errorf("can not get oke workload identity based auth config provider: %w", err)
		}
		return []oci_common.ConfigurationProvider{cfg}, nil
	default:
		return nil, fmt.Errorf("auth must be one of %q, %q, %q, %q, %q, %q", AuthAPIKey, AuthInstancePrincipal, AuthInstanceWithCerts, AuthSecurityToken, AuthResourcePrincipal, AuthOKEWorkloadIdentity)
	}
}

func instancePrincipalProviders(region string) ([]oci_common.ConfigurationProvider, error) {
	authClientModifier := func(client oci_common.HTTPRequestDispatcher) (oci_common.HTTPRequestDispatcher, error) {
		return client, nil
	}

	cfg, err := oci_common_auth.InstancePrincipalConfigurationForRegionWithCustomClient(oci_common.StringToRegion(region), authClientModifier)
	if err != nil {
		return nil, err
	}

	return []oci_common.ConfigurationProvider{cfg}, nil
}

func instancePrincipalWithCertsProviders(d *schema.ResourceData) ([]oci_common.ConfigurationProvider, error) {
	region, ok := d.GetOkExists(RegionAttrName)
	if !ok || region.(string) == "" {
		return nil, fmt.Errorf("can not get %s from Terraform configuration (InstancePrincipalWithCerts)", RegionAttrName)
	}

	defaultCertsDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("can not get working directory for current os platform")
	}

	certsDir := filepath.Clean(envOrDefault(testCertificatesLocation, defaultCertsDir))
	leafCertificateBytes, err := certificateFileBytes(filepath.Join(certsDir, "ip_cert.pem"))
	if err != nil {
		return nil, fmt.Errorf("can not read leaf certificate from %s", filepath.Join(certsDir, "ip_cert.pem"))
	}

	leafPrivateKeyBytes, err := certificateFileBytes(filepath.Join(certsDir, "ip_key.pem"))
	if err != nil {
		return nil, fmt.Errorf("can not read leaf private key from %s", filepath.Join(certsDir, "ip_key.pem"))
	}

	leafPassphraseBytes := []byte{}
	leafPassphrasePath := filepath.Join(certsDir, "leaf_passphrase")
	if _, err := os.Stat(leafPassphrasePath); err == nil {
		leafPassphraseBytes, err = certificateFileBytes(leafPassphrasePath)
		if err != nil {
			return nil, fmt.Errorf("can not read leaf passphrase from %s", leafPassphrasePath)
		}
	}

	intermediateCertificateBytes, err := certificateFileBytes(filepath.Join(certsDir, "intermediate.pem"))
	if err != nil {
		return nil, fmt.Errorf("can not read intermediate certificate from %s", filepath.Join(certsDir, "intermediate.pem"))
	}

	cfg, err := oci_common_auth.InstancePrincipalConfigurationWithCerts(
		oci_common.StringToRegion(region.(string)),
		leafCertificateBytes,
		leafPassphraseBytes,
		leafPrivateKeyBytes,
		[][]byte{intermediateCertificateBytes},
	)
	if err != nil {
		return nil, err
	}

	return []oci_common.ConfigurationProvider{cfg}, nil
}

func securityTokenProviders(d *schema.ResourceData) ([]oci_common.ConfigurationProvider, error) {
	region, ok := d.GetOkExists(RegionAttrName)
	if !ok || region.(string) == "" {
		return nil, fmt.Errorf("can not get %s from Terraform configuration (SecurityToken)", RegionAttrName)
	}

	profile, ok := d.GetOkExists(ConfigFileProfileAttrName)
	if !ok || profile.(string) == "" {
		return nil, fmt.Errorf("missing profile in provider block %s", ConfigFileProfileAttrName)
	}

	privateKeyPassword := d.Get(PrivateKeyPasswordAttrName).(string)
	defaultPath := path.Join(homeFolder(), defaultConfigDirName, defaultConfigFileName)
	if err := checkProfile(profile.(string), defaultPath); err != nil {
		return nil, err
	}

	securityTokenConfigProvider, err := oci_common.ConfigurationProviderForSessionTokenWithProfile(defaultPath, profile.(string), privateKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("could not create security token based auth config provider: %w", err)
	}

	regionProvider := oci_common.NewRawConfigurationProvider("", "", region.(string), "", "", nil)
	return []oci_common.ConfigurationProvider{regionProvider, securityTokenConfigProvider}, nil
}

func resourcePrincipalProviders(d *schema.ResourceData) ([]oci_common.ConfigurationProvider, error) {
	region, ok := d.GetOkExists(RegionAttrName)
	if ok && region.(string) != "" {
		cfg, err := oci_common_auth.ResourcePrincipalConfigurationProviderForRegion(oci_common.StringToRegion(region.(string)))
		if err != nil {
			return nil, err
		}
		return []oci_common.ConfigurationProvider{cfg}, nil
	}

	cfg, err := oci_common_auth.ResourcePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}
	return []oci_common.ConfigurationProvider{cfg}, nil
}

type resourceDataConfigProvider struct {
	d *schema.ResourceData
}

func (p resourceDataConfigProvider) AuthType() (oci_common.AuthConfig, error) {
	return oci_common.AuthConfig{
		AuthType:         oci_common.UnknownAuthenticationType,
		IsFromConfigFile: false,
		OboToken:         nil,
	}, fmt.Errorf("unsupported, keep the interface")
}

func (p resourceDataConfigProvider) TenancyOCID() (string, error) {
	if boatTenancy := envOrDefault(boatTenancyOcidAttrName, ""); boatTenancy != "" {
		return boatTenancy, nil
	}
	if value, ok := p.d.GetOkExists(TenancyOcidAttrName); ok && value.(string) != "" {
		return value.(string), nil
	}
	return "", fmt.Errorf("can not get %s from Terraform configuration", TenancyOcidAttrName)
}

func (p resourceDataConfigProvider) UserOCID() (string, error) {
	if value, ok := p.d.GetOkExists(UserOcidAttrName); ok && value.(string) != "" {
		return value.(string), nil
	}
	return "", fmt.Errorf("can not get %s from Terraform configuration", UserOcidAttrName)
}

func (p resourceDataConfigProvider) KeyFingerprint() (string, error) {
	if value, ok := p.d.GetOkExists(FingerprintAttrName); ok && value.(string) != "" {
		return value.(string), nil
	}
	return "", fmt.Errorf("can not get %s from Terraform configuration", FingerprintAttrName)
}

func (p resourceDataConfigProvider) Region() (string, error) {
	if value, ok := p.d.GetOkExists(RegionAttrName); ok && value.(string) != "" {
		return value.(string), nil
	}
	return "", fmt.Errorf("can not get %s from Terraform configuration", RegionAttrName)
}

func (p resourceDataConfigProvider) KeyID() (string, error) {
	tenancy, err := p.TenancyOCID()
	if err != nil {
		return "", err
	}
	user, err := p.UserOCID()
	if err != nil {
		return "", err
	}
	fingerprint, err := p.KeyFingerprint()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", tenancy, user, fingerprint), nil
}

func (p resourceDataConfigProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	password := ""
	if value, ok := p.d.GetOkExists(PrivateKeyPasswordAttrName); ok {
		password = value.(string)
	}

	if value, ok := p.d.GetOk(PrivateKeyAttrName); ok && value.(string) != "" {
		keyData := strings.ReplaceAll(value.(string), "\\n", "\n")
		return oci_common.PrivateKeyFromBytesWithPassword([]byte(keyData), []byte(password))
	}

	if value, ok := p.d.GetOkExists(PrivateKeyPathAttrName); ok && value.(string) != "" {
		resolvedPath := expandPath(value.(string))
		pemFileContent, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("can not read private key from %q: %w", value.(string), err)
		}
		return oci_common.PrivateKeyFromBytes(pemFileContent, &password)
	}

	return nil, fmt.Errorf("can not get private_key or private_key_path from Terraform configuration")
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
			Proxy:               http.ProxyFromEnvironment,
		},
	}
}

func envOrDefault(name, defaultValue string) string {
	if value := os.Getenv("TF_VAR_" + name); value != "" {
		return value
	}
	if value := os.Getenv("OCI_" + strings.ToUpper(name)); value != "" {
		return value
	}
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

func homeFolder() string {
	if override := os.Getenv("TF_HOME_OVERRIDE"); override != "" {
		return override
	}
	current, err := user.Current()
	if err != nil {
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
		return os.Getenv("USERPROFILE")
	}
	return current.HomeDir
}

func expandPath(filePath string) string {
	if strings.HasPrefix(filePath, fmt.Sprintf("~%c", os.PathSeparator)) {
		filePath = path.Join(homeFolder(), filePath[2:])
	}
	return path.Clean(filePath)
}

func checkProfile(profile, configPath string) error {
	profileRegex := regexp.MustCompile(`^\[(.*)\]`)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		match := profileRegex.FindStringSubmatch(line)
		if len(match) > 1 && match[1] == profile {
			return nil
		}
	}

	return fmt.Errorf("configuration file did not contain profile: %s", profile)
}

func certificateFileBytes(fullPath string) ([]byte, error) {
	absFile, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("can't form absolute path of %s: %w", fullPath, err)
	}

	pemRaw, err := os.ReadFile(absFile)
	if err != nil {
		return nil, fmt.Errorf("can't read %s: %w", fullPath, err)
	}
	return pemRaw, nil
}
