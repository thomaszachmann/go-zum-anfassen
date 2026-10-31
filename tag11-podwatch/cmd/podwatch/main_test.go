package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRunScan(t *testing.T) {
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "shop"},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
			{Name: "app", RestartCount: 4}}},
	}
	tests := []struct {
		name  string
		ns    string
		wantN int
		want  string
	}{
		{"treffer", "shop", 1, "Restarts"},
		{"anderer namespace", "kube-system", 0, "keine auffälligen Pods"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := run(context.Background(), fake.NewClientset(p),
				options{ns: tt.ns, restarts: 3}, &buf)
			if err != nil || n != tt.wantN || !strings.Contains(buf.String(), tt.want) {
				t.Fatalf("n=%d err=%v out=%q", n, err, buf.String())
			}
		})
	}
}
