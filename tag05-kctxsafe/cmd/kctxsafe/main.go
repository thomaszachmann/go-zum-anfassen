// kctxsafe zeigt und wechselt kubeconfig-Kontexte – prod nur mit Bestätigung.
package main

import (
	"context"
	"io"
	"os"
	"os/signal"

	"example.com/kctxsafe/internal/kube"
	"github.com/spf13/cobra"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	root := newRoot(kube.ExecRunner, os.Getenv, os.Stdin)
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

// newRoot baut den Befehlsbaum; Abhängigkeiten kommen von außen (testbar).
func newRoot(run kube.Runner, getenv func(string) string, in io.Reader,
) *cobra.Command {
	a := &app{run: run, getenv: getenv, in: in}
	root := &cobra.Command{
		Use:          "kctxsafe",
		Short:        "kubeconfig-Kontexte anzeigen und sicher wechseln",
		SilenceUsage: true,
	}
	pf := root.PersistentFlags()
	pf.StringVar(&a.kubeconfig, "kubeconfig", "", "Kubeconfig (sonst $KUBECONFIG)")
	pf.StringVar(&a.configFile, "config", "", "Config (sonst $KCTXSAFE_CONFIG)")
	pf.StringVar(&a.prodPattern, "prod-pattern", "", "Regex für prod-Kontexte")
	root.AddCommand(a.listCmd(), a.currentCmd(), a.useCmd())
	return root
}
