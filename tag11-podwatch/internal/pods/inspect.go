// Package pods findet Pods mit CrashLoops, OOMKills und vielen Neustarts.
package pods

import (
	corev1 "k8s.io/api/core/v1"
)

// Finding ist ein auffälliger Container.
type Finding struct {
	Namespace string
	Pod       string
	Container string
	Reason    string
	Restarts  int32
}

// Inspect prüft alle Container eines Pods.
func Inspect(p *corev1.Pod, minRestarts int32) []Finding {
	var out []Finding
	for _, cs := range p.Status.ContainerStatuses {
		reason := ""
		switch {
		case cs.State.Waiting != nil &&
			cs.State.Waiting.Reason == "CrashLoopBackOff":
			reason = "CrashLoopBackOff"
		case cs.LastTerminationState.Terminated != nil &&
			cs.LastTerminationState.Terminated.Reason == "OOMKilled":
			reason = "OOMKilled"
		case cs.RestartCount >= minRestarts:
			reason = "Restarts"
		}
		if reason != "" {
			out = append(out, Finding{p.Namespace, p.Name, cs.Name,
				reason, cs.RestartCount})
		}
	}
	return out
}
