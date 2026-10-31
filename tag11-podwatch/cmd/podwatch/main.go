// podwatch meldet CrashLoops, OOMKills und häufige Neustarts von Pods.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/podwatch/internal/pods"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type options struct {
	ns       string
	restarts int
	watch    bool
}

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Pfad zur kubeconfig")
	kctx := flag.String("context", "", "kubeconfig-Kontext (Default: aktueller)")
	var o options
	flag.StringVar(&o.ns, "n", "", "Namespace (leer = alle)")
	flag.IntVar(&o.restarts, "restarts", 3, "ab so vielen Neustarts melden")
	flag.BoolVar(&o.watch, "watch", false, "dauerhaft beobachten (Informer)")
	flag.Parse()

	cs, err := client(*kubeconfig, *kctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "podwatch:", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()
	n, err := run(ctx, cs, o, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "podwatch:", err)
		os.Exit(2)
	}
	if n > 0 {
		os.Exit(1) // Findings: CI und Skripte sollen das merken
	}
}

// client baut den Clientset wie kubectl: Flag, $KUBECONFIG, ~/.kube/config.
func client(path, kctx string) (kubernetes.Interface, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = path
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
		&clientcmd.ConfigOverrides{CurrentContext: kctx}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	cfg.Timeout = 15 * time.Second
	return kubernetes.NewForConfig(cfg)
}

// run prüft einmalig (Tabelle) oder beobachtet dauerhaft (eine Zeile je Fund).
func run(ctx context.Context, cs kubernetes.Interface, o options,
	w io.Writer) (int, error) {
	if o.watch {
		n := 0
		err := pods.Watch(ctx, cs, o.ns, int32(o.restarts), func(f pods.Finding) {
			n++
			fmt.Fprintf(w, "%s %s/%s %s %s restarts=%d\n",
				time.Now().Format(time.TimeOnly), f.Namespace, f.Pod,
				f.Container, f.Reason, f.Restarts)
		})
		return n, err
	}
	fs, err := pods.Scan(ctx, cs, o.ns, int32(o.restarts))
	if err != nil {
		return 0, err
	}
	if len(fs) == 0 {
		fmt.Fprintln(w, "keine auffälligen Pods")
		return 0, nil
	}
	return len(fs), pods.Print(w, fs)
}
