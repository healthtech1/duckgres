package ducklake

import (
	"fmt"
	"strings"
)

// MetadataSecretName is the DuckDB secret that holds the metadata store's
// credentials when client sessions are restricted.
const MetadataSecretName = "ducklake_metadata"

// metadataSecretFields maps the libpq keywords a DuckDB postgres secret can
// carry to the secret's parameter names. Everything else stays in the ATTACH path.
var metadataSecretFields = map[string]string{
	"host":     "HOST",
	"port":     "PORT",
	"user":     "USER",
	"password": "PASSWORD",
}

type libpqParam struct {
	key   string
	value string
}

// parseLibpqParams parses a libpq keyword/value connection string, including
// single-quoted values with backslash escapes.
func parseLibpqParams(s string) ([]libpqParam, error) {
	var params []libpqParam
	i := 0
	skipSpace := func() {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
	}
	for {
		skipSpace()
		if i >= len(s) {
			return params, nil
		}
		start := i
		for i < len(s) && s[i] != '=' && s[i] != ' ' && s[i] != '\t' {
			i++
		}
		key := s[start:i]
		skipSpace()
		if i >= len(s) || s[i] != '=' || key == "" {
			return nil, fmt.Errorf("malformed connection string: expected key=value near %q", s[start:])
		}
		i++
		skipSpace()
		var value strings.Builder
		if i < len(s) && s[i] == '\'' {
			i++
			closed := false
			for i < len(s) {
				c := s[i]
				if c == '\\' && i+1 < len(s) {
					value.WriteByte(s[i+1])
					i += 2
					continue
				}
				if c == '\'' {
					i++
					closed = true
					break
				}
				value.WriteByte(c)
				i++
			}
			if !closed {
				return nil, fmt.Errorf("malformed connection string: unterminated quote for %q", key)
			}
		} else {
			for i < len(s) && s[i] != ' ' && s[i] != '\t' && s[i] != '\n' && s[i] != '\r' {
				value.WriteByte(s[i])
				i++
			}
		}
		params = append(params, libpqParam{key: key, value: value.String()})
	}
}

func formatLibpqValue(v string) string {
	if v != "" && !strings.ContainsAny(v, " \t\n\r'\\") {
		return v
	}
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(v) + "'"
}

func attachOptions(dlCfg Config, migrate bool) []string {
	var options []string
	dataPath := dlCfg.ObjectStore
	if dataPath == "" {
		dataPath = dlCfg.DataPath
	}
	if dataPath != "" {
		options = append(options, fmt.Sprintf("DATA_PATH '%s'", escapeSQLStringLiteral(dataPath)))
	}
	if migrate {
		options = append(options, "AUTOMATIC_MIGRATION TRUE")
	}
	if dlCfg.DataInliningRowLimit != nil {
		options = append(options, fmt.Sprintf("DATA_INLINING_ROW_LIMIT %d", *dlCfg.DataInliningRowLimit))
	}
	return options
}

// BuildRestrictedAttachStmts splits a "postgres:" metadata store into a
// CREATE SECRET holding its host, port, user and password, and an ATTACH that
// reads them from that secret. DuckDB shows an attached database's path in
// duckdb_databases(), so the password must not be part of it.
func BuildRestrictedAttachStmts(dlCfg Config, migrate bool) (secretStmt, attachStmt string, err error) {
	if !strings.HasPrefix(dlCfg.MetadataStore, "postgres:") {
		return "", "", fmt.Errorf("restricted client sessions need a postgres: metadata store")
	}
	connStr := injectPostgresApplicationName(injectPostgresKeepalive(dlCfg.MetadataStore), dlCfg.ApplicationName)
	params, err := parseLibpqParams(strings.TrimPrefix(connStr, "postgres:"))
	if err != nil {
		return "", "", err
	}

	secretParts := []string{"TYPE postgres"}
	var pathParts []string
	for _, p := range params {
		if field, ok := metadataSecretFields[p.key]; ok {
			secretParts = append(secretParts, fmt.Sprintf("%s '%s'", field, escapeSQLStringLiteral(p.value)))
		} else {
			pathParts = append(pathParts, p.key+"="+formatLibpqValue(p.value))
		}
	}

	secretStmt = fmt.Sprintf("CREATE OR REPLACE SECRET %s (%s)", MetadataSecretName, strings.Join(secretParts, ", "))
	options := append(attachOptions(dlCfg, migrate),
		fmt.Sprintf("METADATA_PARAMETERS MAP {'TYPE': 'postgres', 'SECRET': '%s'}", MetadataSecretName))
	attachStmt = fmt.Sprintf("ATTACH 'ducklake:%s' AS ducklake (%s)",
		escapeSQLStringLiteral(strings.Join(pathParts, " ")), strings.Join(options, ", "))
	return secretStmt, attachStmt, nil
}

// BuildSessionLockStmts returns the statements that stop a client session from
// attaching other databases (including with the metadata secret), reading or
// writing files outside allowedDirs and allowedPaths, and changing any setting
// afterwards. They must run last: once the configuration is locked, nothing can
// SET it again.
func BuildSessionLockStmts(allowedDirs, allowedPaths []string) []string {
	return []string{
		fmt.Sprintf("SET allowed_directories = %s", sqlStringList(allowedDirs)),
		fmt.Sprintf("SET allowed_paths = %s", sqlStringList(allowedPaths)),
		"SET enable_external_access = false",
		"SET lock_configuration = true",
	}
}

func sqlStringList(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = "'" + escapeSQLStringLiteral(v) + "'"
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
