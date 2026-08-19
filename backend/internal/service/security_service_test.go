package service

import "testing"

func TestComputeScore(t *testing.T) {
	cases := []struct {
		name     string
		findings []Finding
		want     int
	}{
		{"clean", nil, 100},
		{"one high", []Finding{{Severity: "high"}}, 90},
		{"mixed", []Finding{{Severity: "high"}, {Severity: "high"}, {Severity: "medium"}, {Severity: "low"}, {Severity: "info"}}, 73},
		{"floor zero", []Finding{{Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}, {Severity: "high"}}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := computeScore(tc.findings); got != tc.want {
				t.Fatalf("computeScore = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCapDropAll(t *testing.T) {
	if !capDropAll([]string{"ALL", "SYS_ADMIN"}) {
		t.Fatal("ALL should satisfy capDropAll")
	}
	if capDropAll([]string{"NET_ADMIN"}) {
		t.Fatal("missing ALL should not satisfy capDropAll")
	}
	if capDropAll(nil) {
		t.Fatal("nil should not satisfy capDropAll")
	}
}
