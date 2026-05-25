package server

import "testing"

func TestCastUntypedParams(t *testing.T) {
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
			"mixed typed and untyped",
			"SELECT * FROM t WHERE a = $1 AND b = $2 AND c = $3",
			[]int32{0, 23, 0}, // $1 untyped, $2 int4, $3 untyped
			"SELECT * FROM t WHERE a = $1::TEXT AND b = $2 AND c = $3::TEXT",
		},
		{
			"all typed",
			"SELECT * FROM t WHERE a = $1 AND b = $2",
			[]int32{25, 23},
			"SELECT * FROM t WHERE a = $1 AND b = $2",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := castUntypedParams(tt.query, tt.paramTypes)
			if got != tt.want {
				t.Errorf("castUntypedParams() =\n  %s\nwant:\n  %s", got, tt.want)
			}
		})
	}
}
