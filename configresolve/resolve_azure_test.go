package configresolve

import (
	"testing"

	"github.com/posthog/duckgres/configloader"
)

func TestResolveEffectiveAzureConfigFromYAML(t *testing.T) {
	fileCfg := &configloader.FileConfig{
		DuckLake: configloader.DuckLakeFileConfig{
			MetadataStore:    "postgres:host=localhost dbname=ducklake",
			ObjectStore:      "azure://data/",
			AzureProvider:    "credential_chain",
			AzureAccountName: "mystorageaccount",
			AzureChain:       "managed_identity",
		},
	}

	resolved := ResolveEffective(fileCfg, CLIInputs{}, nil, nil)

	if resolved.Server.DuckLake.AzureProvider != "credential_chain" {
		t.Fatalf("expected AzureProvider credential_chain, got %q", resolved.Server.DuckLake.AzureProvider)
	}
	if resolved.Server.DuckLake.AzureAccountName != "mystorageaccount" {
		t.Fatalf("expected AzureAccountName mystorageaccount, got %q", resolved.Server.DuckLake.AzureAccountName)
	}
	if resolved.Server.DuckLake.AzureChain != "managed_identity" {
		t.Fatalf("expected AzureChain managed_identity, got %q", resolved.Server.DuckLake.AzureChain)
	}
}

func TestResolveEffectiveAzureConfigFromEnv(t *testing.T) {
	env := map[string]string{
		"DUCKGRES_DUCKLAKE_AZURE_PROVIDER":     "credential_chain",
		"DUCKGRES_DUCKLAKE_AZURE_ACCOUNT_NAME": "envaccount",
		"DUCKGRES_DUCKLAKE_AZURE_CHAIN":        "cli;managed_identity",
	}

	resolved := ResolveEffective(nil, CLIInputs{}, func(key string) string {
		return env[key]
	}, nil)

	if resolved.Server.DuckLake.AzureProvider != "credential_chain" {
		t.Fatalf("expected AzureProvider credential_chain, got %q", resolved.Server.DuckLake.AzureProvider)
	}
	if resolved.Server.DuckLake.AzureAccountName != "envaccount" {
		t.Fatalf("expected AzureAccountName envaccount, got %q", resolved.Server.DuckLake.AzureAccountName)
	}
	if resolved.Server.DuckLake.AzureChain != "cli;managed_identity" {
		t.Fatalf("expected AzureChain cli;managed_identity, got %q", resolved.Server.DuckLake.AzureChain)
	}
}

func TestResolveEffectiveAzureEnvOverridesYAML(t *testing.T) {
	fileCfg := &configloader.FileConfig{
		DuckLake: configloader.DuckLakeFileConfig{
			AzureAccountName: "yaml-account",
		},
	}

	resolved := ResolveEffective(fileCfg, CLIInputs{}, func(key string) string {
		if key == "DUCKGRES_DUCKLAKE_AZURE_ACCOUNT_NAME" {
			return "env-account"
		}
		return ""
	}, nil)

	if resolved.Server.DuckLake.AzureAccountName != "env-account" {
		t.Fatalf("expected env to override YAML, got %q", resolved.Server.DuckLake.AzureAccountName)
	}
}
