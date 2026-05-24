package server

import (
	"strings"
	"testing"
)

func TestIsAzureObjectStore(t *testing.T) {
	tests := []struct {
		name        string
		objectStore string
		want        bool
	}{
		{"azure:// prefix", "azure://data/", true},
		{"az:// prefix", "az://container/path/", true},
		{"s3:// prefix", "s3://bucket/path/", false},
		{"empty string", "", false},
		{"local path", "/tmp/data", false},
		{"azure in path but not scheme", "s3://azure-bucket/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAzureObjectStore(tt.objectStore)
			if got != tt.want {
				t.Errorf("isAzureObjectStore(%q) = %v, want %v", tt.objectStore, got, tt.want)
			}
		})
	}
}

func TestBuildAzureSecret_DefaultProvider(t *testing.T) {
	secret := buildAzureSecret(DuckLakeConfig{
		AzureAccountName: "mystorageaccount",
	})

	if !strings.Contains(secret, "TYPE azure") {
		t.Errorf("expected TYPE azure, got:\n%s", secret)
	}
	if !strings.Contains(secret, "PROVIDER credential_chain") {
		t.Errorf("expected default PROVIDER credential_chain, got:\n%s", secret)
	}
	if !strings.Contains(secret, "ACCOUNT_NAME 'mystorageaccount'") {
		t.Errorf("expected ACCOUNT_NAME 'mystorageaccount', got:\n%s", secret)
	}
	if strings.Contains(secret, "CHAIN") {
		t.Errorf("expected no CHAIN clause when AzureChain is empty, got:\n%s", secret)
	}
}

func TestBuildAzureSecret_ExplicitProvider(t *testing.T) {
	secret := buildAzureSecret(DuckLakeConfig{
		AzureAccountName: "myaccount",
		AzureProvider:    "access_token",
	})

	if !strings.Contains(secret, "PROVIDER access_token") {
		t.Errorf("expected PROVIDER access_token, got:\n%s", secret)
	}
}

func TestBuildAzureSecret_WithChain(t *testing.T) {
	secret := buildAzureSecret(DuckLakeConfig{
		AzureAccountName: "myaccount",
		AzureChain:       "managed_identity;cli",
	})

	if !strings.Contains(secret, "CHAIN 'managed_identity;cli'") {
		t.Errorf("expected CHAIN 'managed_identity;cli', got:\n%s", secret)
	}
}

func TestBuildAzureSecret_WithClientID(t *testing.T) {
	secret := buildAzureSecret(DuckLakeConfig{
		AzureAccountName: "myaccount",
		AzureProvider:    "managed_identity",
		AzureClientID:    "00000000-0000-0000-0000-000000000001",
	})

	if !strings.Contains(secret, "PROVIDER managed_identity") {
		t.Errorf("expected PROVIDER managed_identity, got:\n%s", secret)
	}
	if !strings.Contains(secret, "CLIENT_ID '00000000-0000-0000-0000-000000000001'") {
		t.Errorf("expected CLIENT_ID, got:\n%s", secret)
	}
}

func TestBuildAzureSecret_SecretName(t *testing.T) {
	secret := buildAzureSecret(DuckLakeConfig{
		AzureAccountName: "myaccount",
	})

	if !strings.Contains(secret, "ducklake_azure") {
		t.Errorf("expected secret name ducklake_azure, got:\n%s", secret)
	}
}
