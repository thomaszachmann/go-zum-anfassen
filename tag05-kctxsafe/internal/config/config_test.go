package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReihenfolge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kctxsafe.conf")
	content := "# Lab\nprod_pattern = prod|live\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{}
	getenv := func(k string) string { return env[k] }

	tests := []struct {
		name, path, envVal, want string
	}{
		{"Default", "", "", "prod"},
		{"Datei", path, "", "prod|live"},
		{"Env schlägt Datei", path, "^p-", "^p-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env["KCTXSAFE_PROD_PATTERN"] = tt.envVal
			c, err := Load(tt.path, getenv)
			if err != nil {
				t.Fatal(err)
			}
			if c.ProdPattern != tt.want {
				t.Errorf("ProdPattern = %q, want %q", c.ProdPattern, tt.want)
			}
		})
	}
}
