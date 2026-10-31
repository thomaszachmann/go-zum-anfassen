package lint

import (
	"fmt"
	"slices"
	"strings"
)

// Finding ist ein Verstoß; die JSON-Tags bestimmen die Ausgabe mit -o json.
type Finding struct {
	File    string `json:"file"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// Check wendet alle Regeln auf ein Manifest an.
func Check(m Manifest, requiredLabels []string) []Finding {
	var fs []Finding
	add := func(rule, format string, args ...any) {
		fs = append(fs, Finding{Kind: m.Kind, Name: m.Metadata.Name,
			Rule: rule, Message: fmt.Sprintf(format, args...)})
	}
	for _, l := range requiredLabels {
		if _, ok := m.Metadata.Labels[l]; !ok {
			add("label", "%s fehlt", l)
		}
	}
	pod := m.Pod()
	if pod == nil {
		return fs
	}
	for _, c := range slices.Concat(pod.InitContainers, pod.Containers) {
		if floatingTag(c.Image) {
			add("image-tag", "%s: %s", c.Name, c.Image)
		}
		for _, k := range []string{"cpu", "memory"} {
			if c.Resources.Requests[k] == "" {
				add("resources", "%s: requests.%s fehlt", c.Name, k)
			}
		}
		if c.Resources.Limits["memory"] == "" {
			add("resources", "%s: limits.memory fehlt", c.Name)
		}
	}
	return fs
}

// floatingTag erkennt :latest und fehlende Tags; Digests sind immer fest.
func floatingTag(image string) bool {
	if strings.Contains(image, "@sha256:") {
		return false
	}
	last := image[strings.LastIndex(image, "/")+1:] // Registry-Port ignorieren
	name, tag, ok := strings.Cut(last, ":")
	return name == "" || !ok || tag == "latest"
}
