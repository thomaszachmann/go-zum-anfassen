package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"example.com/kctxsafe/internal/config"
	"example.com/kctxsafe/internal/guard"
	"example.com/kctxsafe/internal/kube"
	"github.com/spf13/cobra"
)

type app struct {
	run                                 kube.Runner
	getenv                              func(string) string
	in                                  io.Reader
	kubeconfig, configFile, prodPattern string
}

// setup wertet Flag > Env > Datei > Default aus.
func (a *app) setup() (kube.Client, *regexp.Regexp, error) {
	path := first(a.kubeconfig, a.getenv("KUBECONFIG"))
	if path == "" {
		return kube.Client{}, nil, errors.New("--kubeconfig oder KUBECONFIG setzen")
	}
	if strings.Contains(path, ":") {
		return kube.Client{}, nil, errors.New("genau eine Kubeconfig-Datei erlaubt")
	}
	cfg, err := config.Load(first(a.configFile, a.getenv("KCTXSAFE_CONFIG")), a.getenv)
	if err != nil {
		return kube.Client{}, nil, err
	}
	if a.prodPattern != "" {
		cfg.ProdPattern = a.prodPattern
	}
	re, err := cfg.ProdRegexp()
	return kube.Client{Kubeconfig: path, Run: a.run}, re, err
}

func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (a *app) listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Kontexte auflisten (* = aktiv)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, prod, err := a.setup()
			if err != nil {
				return err
			}
			names, err := c.Contexts(cmd.Context())
			if err != nil {
				return err
			}
			cur, _ := c.Current(cmd.Context()) // kein aktiver Kontext ist ok
			for _, n := range names {
				mark, tag := " ", ""
				if n == cur {
					mark = "*"
				}
				if prod.MatchString(n) {
					tag = "  [PROD]"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s%s\n", mark, n, tag)
			}
			return nil
		},
	}
}

func (a *app) currentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "aktiven Kontext anzeigen",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := a.setup()
			if err != nil {
				return err
			}
			cur, err := c.Current(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), cur)
			return nil
		},
	}
}

func (a *app) useCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "use KONTEXT",
		Short: "Kontext wechseln (prod nur mit Bestätigung)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, prod, err := a.setup()
			if err != nil {
				return err
			}
			name := args[0]
			if prod.MatchString(name) && !yes {
				ok, err := guard.Confirm(a.in, cmd.ErrOrStderr(), name)
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("abgebrochen, Kontext unverändert")
				}
			}
			if err := c.Use(cmd.Context(), name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Kontext jetzt: %s\n", name)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "ohne Rückfrage (für Skripte)")
	return cmd
}
