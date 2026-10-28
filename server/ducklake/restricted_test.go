package ducklake

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseLibpqParams(t *testing.T) {
	got, err := parseLibpqParams(`host=pg.example.com  port = 5432 password='it\'s \'quoted\' \\ ok' dbname=ducklake`)
	if err != nil {
		t.Fatalf("parseLibpqParams: %v", err)
	}
	want := []libpqParam{
		{key: "host", value: "pg.example.com"},
		{key: "port", value: "5432"},
		{key: "password", value: `it's 'quoted' \ ok`},
		{key: "dbname", value: "ducklake"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseLibpqParams = %#v, want %#v", got, want)
	}
}

func TestParseLibpqParamsRejectsMalformedInput(t *testing.T) {
	for _, input := range []string{"host", "password='unterminated", "=value"} {
		if _, err := parseLibpqParams(input); err == nil {
			t.Errorf("parseLibpqParams(%q) succeeded, want an error", input)
		}
	}
}

func TestBuildRestrictedAttachStmtsKeepsCredentialsOutOfThePath(t *testing.T) {
	secretStmt, attachStmt, err := BuildRestrictedAttachStmts(Config{
		MetadataStore:   "postgres:host=pg.example.com port=5432 dbname=ducklake user=duckgres password=s3cr'et sslmode=require",
		ObjectStore:     "azure://data/",
		ApplicationName: "duckgres/test",
	}, false)
	if err != nil {
		t.Fatalf("BuildRestrictedAttachStmts: %v", err)
	}

	wantSecret := "CREATE OR REPLACE SECRET ducklake_metadata (TYPE postgres, HOST 'pg.example.com', PORT '5432', USER 'duckgres', PASSWORD 's3cr''et')"
	if secretStmt != wantSecret {
		t.Errorf("secret statement:\n got %s\nwant %s", secretStmt, wantSecret)
	}

	wantAttach := "ATTACH 'ducklake:dbname=ducklake sslmode=require keepalives=1 keepalives_idle=60 keepalives_interval=10 keepalives_count=5 application_name=duckgres/test' AS ducklake " +
		"(DATA_PATH 'azure://data/', METADATA_PARAMETERS MAP {'TYPE': 'postgres', 'SECRET': 'ducklake_metadata'})"
	if attachStmt != wantAttach {
		t.Errorf("attach statement:\n got %s\nwant %s", attachStmt, wantAttach)
	}
	for _, leaked := range []string{"s3cr", "pg.example.com", "duckgres'", "user="} {
		if strings.Contains(attachStmt, leaked) {
			t.Errorf("attach statement contains %q: %s", leaked, attachStmt)
		}
	}
}

func TestBuildRestrictedAttachStmtsRequiresPostgres(t *testing.T) {
	if _, _, err := BuildRestrictedAttachStmts(Config{MetadataStore: "sqlite:/tmp/meta.db"}, false); err == nil {
		t.Fatal("BuildRestrictedAttachStmts succeeded for a sqlite metadata store, want an error")
	}
}

func TestBuildSessionLockStmts(t *testing.T) {
	got := BuildSessionLockStmts(
		[]string{"/data/tmp/", "azure://data/", "/it's/"},
		[]string{"/tmp/profile.json"},
	)
	want := []string{
		"SET allowed_directories = ['/data/tmp/', 'azure://data/', '/it''s/']",
		"SET allowed_paths = ['/tmp/profile.json']",
		"SET enable_external_access = false",
		"SET lock_configuration = true",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildSessionLockStmts = %#v, want %#v", got, want)
	}
}
