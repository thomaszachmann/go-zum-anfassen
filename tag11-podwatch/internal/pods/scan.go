package pods

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Scan listet Pods einmalig ("" = alle Namespaces) und prüft sie.
func Scan(ctx context.Context, cs kubernetes.Interface, ns string,
	minRestarts int32) ([]Finding, error) {
	list, err := cs.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pods auflisten: %w", err)
	}
	var out []Finding
	for i := range list.Items {
		out = append(out, Inspect(&list.Items[i], minRestarts)...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Restarts > out[j].Restarts })
	return out, nil
}

// Print schreibt die Findings als ausgerichtete Tabelle.
func Print(w io.Writer, fs []Finding) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tPOD\tCONTAINER\tGRUND\tRESTARTS")
	for _, f := range fs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
			f.Namespace, f.Pod, f.Container, f.Reason, f.Restarts)
	}
	return tw.Flush()
}
