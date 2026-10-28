package server

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
)

func openRestrictedDB(t *testing.T, restrict bool) (*sql.DB, string, string) {
	t.Helper()
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, stmt := range ProfilingSetupSQL(ProfilingOutputPath) {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	dataDir := t.TempDir()
	lakeDir := t.TempDir() + string(filepath.Separator)
	cfg := Config{DataDir: dataDir, DuckLake: DuckLakeConfig{RestrictClientSessions: restrict, DataPath: lakeDir}}
	if err := restrictClientSession(db, cfg); err != nil {
		t.Fatalf("restrictClientSession: %v", err)
	}
	return db, dataDir, lakeDir
}

func TestRestrictClientSessionBlocksAttachFilesAndSettings(t *testing.T) {
	db, _, lakeDir := openRestrictedDB(t, true)

	if _, err := db.Exec(fmt.Sprintf("COPY (SELECT 1 AS id) TO '%sok.parquet'", lakeDir)); err != nil {
		t.Fatalf("writing inside the lake's data path should still work: %v", err)
	}
	var id int
	if err := db.QueryRow(fmt.Sprintf("SELECT id FROM '%sok.parquet'", lakeDir)).Scan(&id); err != nil || id != 1 {
		t.Fatalf("reading inside the lake's data path should still work: id=%d err=%v", id, err)
	}

	blocked := map[string]string{
		"attach another database": fmt.Sprintf("ATTACH '%s' AS other", filepath.Join(t.TempDir(), "other.duckdb")),
		"read a file elsewhere":   "SELECT * FROM read_text('/etc/hostname')",
		"re-enable file access":   "SET enable_external_access = true",
		"unlock the settings":     "SET lock_configuration = false",
		"reveal secrets":          "SET allow_unredacted_secrets = true",
	}
	for name, stmt := range blocked {
		if _, err := db.Exec(stmt); err == nil {
			t.Errorf("%s: %q succeeded, want it refused", name, stmt)
		} else if !strings.Contains(err.Error(), "disabled") && !strings.Contains(err.Error(), "locked") {
			t.Errorf("%s: %q failed for an unexpected reason: %v", name, stmt, err)
		}
	}
}

func TestRestrictClientSessionDoesNothingWhenOff(t *testing.T) {
	db, _, _ := openRestrictedDB(t, false)

	if _, err := db.Exec("SET memory_limit = '1GB'"); err != nil {
		t.Fatalf("settings should stay changeable when RestrictClientSessions is off: %v", err)
	}
}
