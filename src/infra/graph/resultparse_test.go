package graph

import "testing"

func TestParseCompact(t *testing.T) {
	t.Run("parses_rows_and_skips_bad_rows", func(t *testing.T) {
		resp := []any{
			[]any{
				[]any{int64(0), "from_type"},
				[]any{int64(0), "from_key"},
			},
			[]any{
				[]any{[]any{int64(0), "USER"}, []any{int64(0), "u1"}},
				"bad_row",
			},
			[]any{"stats"},
		}

		rows, err := ParseCompact(resp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if rows[0]["from_type"] != "USER" || rows[0]["from_key"] != "u1" {
			t.Fatalf("unexpected row: %#v", rows[0])
		}
	})
}

func TestParseCompactErrors(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"not_array", "nope"},
		{"bad_header", []any{"header", []any{}}},
		{"bad_rows", []any{[]any{}, "rows"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseCompact(tc.in); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}
