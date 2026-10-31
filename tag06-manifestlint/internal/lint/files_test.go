package lint

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFiles(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"a.yaml", "sub/b.yml", "README.md", ".git/c.yaml"} {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Files([]string{dir, filepath.Join(dir, "README.md"), "-"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(dir, "a.yaml"), filepath.Join(dir, "sub/b.yml"),
		filepath.Join(dir, "README.md"), "-", // explizit genannt: immer dabei
	}
	if !slices.Equal(got, want) {
		t.Errorf("Files =\n%v\nwant\n%v", got, want)
	}
}
