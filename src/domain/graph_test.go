package domain

import "testing"

func TestValidateEdgeType(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"trims_and_uppercased", " payment_1 ", "PAYMENT_1", false},
		{"empty", "", "", true},
		{"invalid_chars", "bad type!", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateEdgeType(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected edge type: %s", got)
			}
		})
	}
}

func TestNodeSpecForType(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		label   string
		keyProp string
		node    string
		ok      bool
	}{
		{"merchant", "MERCHANT", "Merchant", "merchant_id_mpan", "MERCHANT", true},
		{"unknown", "UNKNOWN", "", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			label, keyProp, nodeType, ok := nodeSpecForType(tc.in)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: %v", ok)
			}
			if label != tc.label || keyProp != tc.keyProp || string(nodeType) != tc.node {
				t.Fatalf("unexpected spec: %s %s %s", label, keyProp, nodeType)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want int64
		ok   bool
	}{
		{"int64", int64(5), 5, true},
		{"int", int(6), 6, true},
		{"float64", float64(7), 7, true},
		{"string", "8", 8, true},
		{"bad", "nope", 0, false},
		{"unsupported", struct{}{}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toInt64(tc.in)
			if ok != tc.ok {
				t.Fatalf("unexpected ok: %v", ok)
			}
			if ok && got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
