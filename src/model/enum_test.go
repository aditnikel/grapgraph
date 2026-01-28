package model

import "testing"

func TestParseEventType(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    EventType
		wantErr bool
	}{
		{"trims_and_uppercased", " payment ", "PAYMENT", false},
		{"empty", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseEventType(tc.in)
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
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestIsMoneyBearing(t *testing.T) {
	cases := []struct {
		name string
		in   EventType
		want bool
	}{
		{"money_bearing_keyword", EventType("transaction_in"), true},
		{"non_money_event", EventType("login"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsMoneyBearing(tc.in); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
