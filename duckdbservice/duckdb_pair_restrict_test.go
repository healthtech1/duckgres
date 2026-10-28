package duckdbservice

import (
	"testing"
	"time"

	"github.com/posthog/duckgres/server"
)

func TestCreateWorkerDBPairIgnoresRestrictClientSessions(t *testing.T) {
	cfg := server.Config{DataDir: t.TempDir(), DuckLake: server.DuckLakeConfig{RestrictClientSessions: true}}
	pair, err := CreateWorkerDBPair(cfg, make(chan struct{}, 1), "test", time.Now(), "test")
	if err != nil {
		t.Fatalf("CreateWorkerDBPair: %v", err)
	}
	t.Cleanup(func() { _ = pair.Close() })

	if _, err := pair.Main.Exec("SET memory_limit = '1GB'"); err != nil {
		t.Fatalf("workers must stay resizable: %v", err)
	}
}
