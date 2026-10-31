package pods

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// pod baut einen Test-Pod mit einem Container "app".
func pod(name string, restarts int32, waiting, lastTerm string) *corev1.Pod {
	cs := corev1.ContainerStatus{Name: "app", RestartCount: restarts}
	if waiting != "" {
		cs.State.Waiting = &corev1.ContainerStateWaiting{Reason: waiting}
	}
	if lastTerm != "" {
		cs.LastTerminationState.Terminated =
			&corev1.ContainerStateTerminated{Reason: lastTerm}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "shop"},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{cs}},
	}
}

func TestInspect(t *testing.T) {
	tests := []struct {
		name string
		pod  *corev1.Pod
		want string // "" = kein Finding
	}{
		{"gesund", pod("web", 0, "", ""), ""},
		{"crashloop", pod("api", 7, "CrashLoopBackOff", "Error"),
			"CrashLoopBackOff"},
		{"oom", pod("db", 1, "", "OOMKilled"), "OOMKilled"},
		{"viele restarts", pod("job", 5, "", "Completed"), "Restarts"},
		{"knapp darunter", pod("job", 2, "", "Completed"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Inspect(tt.pod, 3)
			if tt.want == "" {
				if len(got) != 0 {
					t.Fatalf("unerwartet: %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0].Reason != tt.want {
				t.Fatalf("got %+v, want Reason %q", got, tt.want)
			}
		})
	}
}
