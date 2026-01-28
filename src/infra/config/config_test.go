package config

import "testing"

func TestSplitCSV(t *testing.T) {
	t.Run("trims_and_skips_empty", func(t *testing.T) {
		got := splitCSV("a, b, ,c,, ")
		if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
			t.Fatalf("unexpected split: %#v", got)
		}
	})
}

func TestEnvHelpers(t *testing.T) {
	t.Run("envStr_trims", func(t *testing.T) {
		t.Setenv("STR_KEY", " value ")
		if got := envStr("STR_KEY", "default"); got != "value" {
			t.Fatalf("unexpected envStr: %s", got)
		}
	})
	t.Run("envInt_parses", func(t *testing.T) {
		t.Setenv("INT_KEY", "10")
		if got := envInt("INT_KEY", 5); got != 10 {
			t.Fatalf("unexpected envInt: %d", got)
		}
	})
	t.Run("envInt_falls_back_on_bad_value", func(t *testing.T) {
		t.Setenv("INT_KEY_BAD", "nope")
		if got := envInt("INT_KEY_BAD", 5); got != 5 {
			t.Fatalf("expected default for bad int, got %d", got)
		}
	})
}

func TestLoadDefaultsAndValidation(t *testing.T) {
	t.Run("errors_on_empty_addrs", func(t *testing.T) {
		t.Setenv("REDIS_ADDRS", " , ")
		if _, err := Load(); err == nil {
			t.Fatalf("expected error when REDIS_ADDRS is empty")
		}
	})
	t.Run("clamps_defaults_and_rank_fallback", func(t *testing.T) {
		t.Setenv("REDIS_ADDRS", "localhost:6379")
		t.Setenv("DEFAULT_MAX_NODES", "0")
		t.Setenv("DEFAULT_MAX_EDGES", "-1")
		t.Setenv("DEFAULT_MIN_EVENT_COUNT", "0")
		t.Setenv("DEFAULT_RANK_BY", "bogus")
		c, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.DefaultMaxNodes <= 0 || c.DefaultMaxEdges <= 0 || c.DefaultMinEventCount <= 0 {
			t.Fatalf("expected defaults to be clamped to positive values")
		}
		if c.DefaultRankBy != "event_count_30d" {
			t.Fatalf("expected default rank fallback, got %s", c.DefaultRankBy)
		}
	})
}
