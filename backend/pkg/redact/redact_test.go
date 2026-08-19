package redact

import (
	"reflect"
	"testing"
)

func TestMaskSensitiveKeys(t *testing.T) {
	in := map[string]any{
		"password":   "P@ssw0rd",
		"token":      "abc123",
		"user_name":  "alice",
		"APIToken":   "xyz",
		"DB_PASSWD":  "secret",
		"nested":     map[string]any{"Authorization": "Bearer xx", "name": "ok"},
		"my_secret":  "hidden",
		"plain_key":  "visible",
	}
	out := Map(in)
	expect := map[string]any{
		"password":  "***",
		"token":     "***",
		"user_name": "alice",
		"APIToken":  "***",
		"DB_PASSWD": "***",
		"nested":    map[string]any{"Authorization": "***", "name": "ok"},
		"my_secret": "***",
		"plain_key": "visible",
	}
	if !reflect.DeepEqual(out, expect) {
		t.Fatalf("mismatch:\n got=%#v\nwant=%#v", out, expect)
	}
}

func TestNonMapPassthrough(t *testing.T) {
	if Map(map[string]any{"a": 1})["a"] != 1 {
		t.Fatal("non-sensitive scalar should pass through")
	}
}

func TestEmptyMap(t *testing.T) {
	if out := Map(map[string]any{}); len(out) != 0 {
		t.Fatal("empty map should stay empty")
	}
}
