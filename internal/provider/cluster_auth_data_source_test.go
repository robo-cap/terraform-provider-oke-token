// Copyright (c) 2026
// Licensed under the Mozilla Public License v2.0

package provider

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fakeSigner struct{}

func (fakeSigner) Sign(r *http.Request) error {
	r.Header.Set("authorization", "Signature version=\"1\",keyId=\"test\"")
	return nil
}

func TestGenerateClusterAuthToken(t *testing.T) {
	createdAt := time.Date(2026, time.April, 2, 10, 30, 0, 0, time.UTC)

	result, err := generateClusterAuthToken(fakeSigner{}, "us-ashburn-1", "ocid1.cluster.oc1..example", createdAt)
	if err != nil {
		t.Fatalf("generateClusterAuthToken returned error: %v", err)
	}

	if got, want := result.Expiration, createdAt.Add(accessTokenExpiration).Format(http.TimeFormat); got != want {
		t.Fatalf("unexpected expiration: got %q want %q", got, want)
	}

	decoded, err := base64.URLEncoding.DecodeString(result.Token)
	if err != nil {
		t.Fatalf("token is not valid base64: %v", err)
	}

	parsedURL, err := url.Parse(string(decoded))
	if err != nil {
		t.Fatalf("decoded token is not a URL: %v", err)
	}

	if got, want := parsedURL.Host, "containerengine.us-ashburn-1.oraclecloud.com"; got != want {
		t.Fatalf("unexpected URL host: got %q want %q", got, want)
	}

	if got, want := parsedURL.Path, "/cluster_request/ocid1.cluster.oc1..example"; got != want {
		t.Fatalf("unexpected URL path: got %q want %q", got, want)
	}

	query := parsedURL.Query()
	if got, want := query.Get("date"), createdAt.Format(http.TimeFormat); got != want {
		t.Fatalf("unexpected date query param: got %q want %q", got, want)
	}

	if auth := query.Get("authorization"); !strings.Contains(auth, "keyId=\"test\"") {
		t.Fatalf("authorization query param missing signer output: %q", auth)
	}
}
