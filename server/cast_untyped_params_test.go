package server

import "testing"

func TestCastParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		paramTypes []int32
		want       string
	}{
		{
			"no params",
			"SELECT 1",
			nil,
			"SELECT 1",
		},
		{
			"all untyped",
			"SELECT * FROM t WHERE a = $1 AND b = $2",
			[]int32{0, 0},
			"SELECT * FROM t WHERE a = $1::TEXT AND b = $2::TEXT",
		},
		{
			"typed int4 and bigint",
			"SELECT * FROM t WHERE a = $1 AND b = $2",
			[]int32{23, 20}, // int4, int8
			"SELECT * FROM t WHERE a = $1::INTEGER AND b = $2::BIGINT",
		},
		{
			"typed oid",
			"SELECT * FROM pg_type WHERE oid = $1",
			[]int32{20}, // int8 (JDBC sends OID as bigint)
			"SELECT * FROM pg_type WHERE oid = $1::BIGINT",
		},
		{
			"mixed typed and untyped",
			"SELECT * FROM t WHERE a = $1 AND b = $2 AND c = $3",
			[]int32{0, 23, 0}, // $1 untyped, $2 int4, $3 untyped
			"SELECT * FROM t WHERE a = $1::TEXT AND b = $2::INTEGER AND c = $3::TEXT",
		},
		{
			"varchar and timestamptz",
			"SELECT * FROM t WHERE name = $1 AND created_at > $2",
			[]int32{1043, 1184},
			"SELECT * FROM t WHERE name = $1::VARCHAR AND created_at > $2::TIMESTAMPTZ",
		},
		{
			"unknown oid falls back to TEXT",
			"SELECT * FROM t WHERE a = $1",
			[]int32{99999},
			"SELECT * FROM t WHERE a = $1::TEXT",
		},
		{
			"double-digit params",
			"SELECT * FROM t WHERE a = $1 AND b = $10",
			[]int32{0, 23, 23, 23, 23, 23, 23, 23, 23, 0},
			"SELECT * FROM t WHERE a = $1::TEXT AND b = $10::TEXT",
		},
		{
			"param in subquery",
			"SELECT * FROM (SELECT * FROM t WHERE x = $1) sub WHERE y = $2",
			[]int32{0, 0},
			"SELECT * FROM (SELECT * FROM t WHERE x = $1::TEXT) sub WHERE y = $2::TEXT",
		},
		{
			"empty param types",
			"SELECT $1",
			[]int32{},
			"SELECT $1",
		},
		{
			"boolean param",
			"SELECT * FROM t WHERE active = $1",
			[]int32{16},
			"SELECT * FROM t WHERE active = $1::BOOLEAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := castParams(tt.query, tt.paramTypes)
			if got != tt.want {
				t.Errorf("castParams() =\n  %s\nwant:\n  %s", got, tt.want)
			}
		})
	}
}
