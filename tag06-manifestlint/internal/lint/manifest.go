// Package lint prüft Kubernetes-Manifeste auf typische Produktionsfehler.
package lint

import (
	"errors"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

// Manifest enthält nur die Felder, die die Regeln brauchen.
type Manifest struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name   string            `yaml:"name"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		PodSpec  `yaml:",inline"` // Pod
		Template struct {
			Spec PodSpec `yaml:"spec"`
		} `yaml:"template"` // Deployment, StatefulSet, DaemonSet, Job
		JobTemplate struct {
			Spec struct {
				Template struct {
					Spec PodSpec `yaml:"spec"`
				} `yaml:"template"`
			} `yaml:"spec"`
		} `yaml:"jobTemplate"` // CronJob
	} `yaml:"spec"`
}

// PodSpec ist der Teil, in dem Container stehen.
type PodSpec struct {
	Containers     []Container `yaml:"containers"`
	InitContainers []Container `yaml:"initContainers"`
}

// Container mit Image und Ressourcen.
type Container struct {
	Name      string `yaml:"name"`
	Image     string `yaml:"image"`
	Resources struct {
		Requests map[string]string `yaml:"requests"`
		Limits   map[string]string `yaml:"limits"`
	} `yaml:"resources"`
}

// Pod liefert die PodSpec je nach Kind, sonst nil.
func (m *Manifest) Pod() *PodSpec {
	switch m.Kind {
	case "Pod":
		return &m.Spec.PodSpec
	case "Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "Job":
		return &m.Spec.Template.Spec
	case "CronJob":
		return &m.Spec.JobTemplate.Spec.Template.Spec
	}
	return nil
}

// Decode liest alle YAML-Dokumente (getrennt durch ---) aus r.
func Decode(r io.Reader) ([]Manifest, error) {
	dec := yaml.NewDecoder(r)
	var out []Manifest
	for i := 1; ; i++ {
		var m Manifest
		err := dec.Decode(&m)
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return out, fmt.Errorf("%d. Dokument: %w", i, err)
		}
		if m.Kind != "" { // leere Dokumente überspringen
			out = append(out, m)
		}
	}
}
