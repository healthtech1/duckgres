package configresolve

import (
	"testing"

	"github.com/posthog/duckgres/configloader"
)

func TestResolveEffectiveRestrictClientSessionsDefaultsOff(t *testing.T) {
	resolved := ResolveEffective(nil, CLIInputs{}, nil, nil)

	if resolved.Server.DuckLake.RestrictClientSessions {
		t.Fatal("expected RestrictClientSessions to default to false")
	}
}

func TestResolveEffectiveRestrictClientSessionsFromYAML(t *testing.T) {
	enabled := true
	fileCfg := &configloader.FileConfig{
		DuckLake: configloader.DuckLakeFileConfig{RestrictClientSessions: &enabled},
	}

	resolved := ResolveEffective(fileCfg, CLIInputs{}, nil, nil)

	if !resolved.Server.DuckLake.RestrictClientSessions {
		t.Fatal("expected restrict_client_sessions: true in YAML to enable RestrictClientSessions")
	}
}

func TestResolveEffectiveRestrictClientSessionsEnvOverridesYAML(t *testing.T) {
	enabled := true
	fileCfg := &configloader.FileConfig{
		DuckLake: configloader.DuckLakeFileConfig{RestrictClientSessions: &enabled},
	}

	resolved := ResolveEffective(fileCfg, CLIInputs{}, func(key string) string {
		if key == "DUCKGRES_DUCKLAKE_RESTRICT_CLIENT_SESSIONS" {
			return "false"
		}
		return ""
	}, nil)

	if resolved.Server.DuckLake.RestrictClientSessions {
		t.Fatal("expected DUCKGRES_DUCKLAKE_RESTRICT_CLIENT_SESSIONS=false to override YAML")
	}
}
