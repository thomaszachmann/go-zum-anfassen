package semver

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    Version
		wantErr bool
	}{
		{in: "1.2.3", want: Version{Major: 1, Minor: 2, Patch: 3}},
		{in: "v0.10.0", want: Version{Minor: 10, V: true}},
		{in: "1.0.0-rc.1", want: Version{Major: 1, Pre: []string{"rc", "1"}}},
		{in: "1.0.0+git.abc", want: Version{Major: 1, Build: "git.abc"}},
		{in: "1.2", wantErr: true},
		{in: "1.02.3", wantErr: true},
		{in: "1.2.3-", wantErr: true},
		{in: "1.2.3-rc..1", wantErr: true},
		{in: "1.2.3-01", wantErr: true},
		{in: "1.2.3+", wantErr: true},
		{in: "x.y.z", wantErr: true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if tt.wantErr {
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("Parse(%q) err = %v, will ErrInvalid", tt.in, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("Parse(%q) unerwarteter Fehler: %v", tt.in, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Parse(%q)\n got: %+v\nwant: %+v", tt.in, got, tt.want)
		}
	}
}
