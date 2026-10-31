package main

import "testing"

func TestParseTriage(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{`{"severity":"high","summary":"Disk voll","action":"aufräumen"}`, false},
		{`{"severity":"kritisch","summary":"x","action":"y"}`, true},
		{"Klar, hier ist dein JSON: {", true},
	}
	for _, tt := range tests {
		if _, err := parseTriage(tt.in); (err != nil) != tt.wantErr {
			t.Errorf("%q: err=%v, wantErr=%v", tt.in, err, tt.wantErr)
		}
	}
}
