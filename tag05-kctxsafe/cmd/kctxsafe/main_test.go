package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"example.com/kctxsafe/internal/kube"
)

// fakeKubectl antwortet wie kubectl und merkt sich, ob gewechselt wurde.
type fakeKubectl struct{ switched bool }

func (f *fakeKubectl) run(_ context.Context, _ string, args ...string) ([]byte, error) {
	switch args[1] {
	case "get-contexts":
		return []byte("kind-dev\nprod-eu\n"), nil
	case "current-context":
		return []byte("kind-dev\n"), nil
	case "use-context":
		f.switched = true
	}
	return nil, nil
}

// execute startet das CLI wie von der Shell, nur ohne Prozess.
func execute(t *testing.T, run kube.Runner, kc, stdin string, args ...string) (
	string, error) {
	t.Helper()
	root := newRoot(run, func(string) string { return "" }, strings.NewReader(stdin))
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"--kubeconfig", kc}, args...))
	err := root.Execute()
	return out.String(), err
}

func labKubeconfig(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("../../lab-kubeconfig.yaml")
	if err != nil {
		t.Fatal(err)
	}
	kc := filepath.Join(t.TempDir(), "kubeconfig") // Kopie, Original bleibt
	if err := os.WriteFile(kc, src, 0o600); err != nil {
		t.Fatal(err)
	}
	return kc
}

func TestUse(t *testing.T) {
	tests := []struct {
		name, stdin string
		args        []string
		wantSwitch  bool
	}{
		{"dev ohne Frage", "", []string{"use", "kind-dev"}, true},
		{"prod mit y", "y\n", []string{"use", "prod-eu"}, false},
		{"prod mit Namen", "prod-eu\n", []string{"use", "prod-eu"}, true},
		{"prod mit --yes", "", []string{"use", "prod-eu", "--yes"}, true},
	}
	kc := labKubeconfig(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeKubectl{}
			_, err := execute(t, f.run, kc, tt.stdin, tt.args...)
			if f.switched != tt.wantSwitch || (err == nil) != tt.wantSwitch {
				t.Errorf("switched=%v err=%v, want switch %v",
					f.switched, err, tt.wantSwitch)
			}
		})
	}
}

func TestListMarkiert(t *testing.T) {
	out, err := execute(t, (&fakeKubectl{}).run, labKubeconfig(t), "", "list")
	if want := "* kind-dev\n  prod-eu  [PROD]\n"; err != nil || out != want {
		t.Errorf("list = %q, %v; want %q", out, err, want)
	}
}

// Integration: echtes kubectl auf einer Kopie der Lab-Kubeconfig.
func TestEchtesKubectl(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl nicht installiert")
	}
	kc := labKubeconfig(t)
	if _, err := execute(t, kube.ExecRunner, kc, "", "use", "staging"); err != nil {
		t.Fatal(err)
	}
	if out, _ := execute(t, kube.ExecRunner, kc, "", "current"); out != "staging\n" {
		t.Errorf("current = %q, want staging", out)
	}
}
