package pods

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cs := fake.NewClientset(pod("api", 7, "CrashLoopBackOff", "Error"))
	// Der Fake kennt keine resourceVersion: erst ändern, wenn Watch läuft.
	started := make(chan struct{})
	cs.PrependWatchReactor("pods", func(a k8stesting.Action) (
		bool, watch.Interface, error) {
		w, err := cs.Tracker().Watch(a.GetResource(), a.GetNamespace())
		close(started)
		return true, w, err
	})
	got := make(chan Finding, 10)
	go func() { _ = Watch(ctx, cs, "shop", 3, func(f Finding) { got <- f }) }()

	expect := func(reason string) {
		t.Helper()
		select {
		case f := <-got:
			if f.Reason != reason {
				t.Fatalf("got %+v, want %s", f, reason)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("kein Finding %s", reason)
		}
	}
	expect("CrashLoopBackOff") // aus der initialen Liste
	<-started
	_, err := cs.CoreV1().Pods("shop").Update(ctx,
		pod("api", 8, "", "OOMKilled"), metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	expect("OOMKilled") // aus dem Watch-Event
}
