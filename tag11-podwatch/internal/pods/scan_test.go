package pods

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestScanAndPrint(t *testing.T) {
	cs := fake.NewClientset(
		pod("web", 0, "", ""),
		pod("api", 7, "CrashLoopBackOff", "Error"),
		pod("db", 1, "", "OOMKilled"),
	)
	fs, err := Scan(context.Background(), cs, "", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 2 || fs[0].Pod != "api" {
		t.Fatalf("got %+v", fs)
	}
	var buf bytes.Buffer
	if err := Print(&buf, fs); err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + buf.String())
	if !strings.Contains(buf.String(), "shop       api  app        CrashLoopBackOff") {
		t.Errorf("Tabelle passt nicht:\n%s", buf.String())
	}
}
