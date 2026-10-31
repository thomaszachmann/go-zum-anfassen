package lint

import (
	"strings"
	"testing"
)

func TestFloatingTag(t *testing.T) {
	tests := map[string]bool{
		"nginx":                          true,
		"nginx:latest":                   true,
		"nginx:1.27":                     false,
		"registry.local:5000/app":        true,
		"registry.local:5000/app:1.0":    false,
		"ghcr.io/x/app@sha256:0b8a3f6c2": false,
	}
	for image, want := range tests {
		t.Run(image, func(t *testing.T) {
			if got := floatingTag(image); got != want {
				t.Errorf("floatingTag(%q) = %v, want %v", image, got, want)
			}
		})
	}
}

const multiDoc = `
kind: Pod
metadata: {name: ok, labels: {app.kubernetes.io/name: ok}}
spec:
  containers:
  - name: app
    image: nginx:1.27
    resources: {requests: {cpu: 1, memory: 1Gi}, limits: {memory: 1Gi}}
---
# leeres Dokument
---
kind: CronJob
metadata: {name: nightly}
spec:
  jobTemplate: {spec: {template: {spec: {containers: [{name: job, image: x}]}}}}
`

func TestDecodeUndCheck(t *testing.T) {
	ms, err := Decode(strings.NewReader(multiDoc))
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 2 {
		t.Fatalf("%d Manifeste, want 2", len(ms))
	}
	if fs := Check(ms[0], []string{"app.kubernetes.io/name"}); len(fs) != 0 {
		t.Errorf("Pod: unerwartete Befunde %v", fs)
	}
	var rules []string
	for _, f := range Check(ms[1], []string{"app.kubernetes.io/name"}) {
		rules = append(rules, f.Rule)
	}
	want := "label image-tag resources resources resources"
	if got := strings.Join(rules, " "); got != want {
		t.Errorf("CronJob: Regeln %q, want %q", got, want)
	}
}

func TestDecodeFehler(t *testing.T) {
	_, err := Decode(strings.NewReader("kind: Pod\n---\nmetadata: [kaputt\n"))
	if err == nil || !strings.Contains(err.Error(), "2. Dokument") {
		t.Errorf("err = %v, want Hinweis auf 2. Dokument", err)
	}
}
