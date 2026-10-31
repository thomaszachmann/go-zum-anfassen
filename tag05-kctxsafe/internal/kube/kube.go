// Package kube ruft kubectl auf – immer mit expliziter --kubeconfig.
package kube

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner führt ein Programm aus und liefert stdout. In Tests ersetzt du ihn.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// ExecRunner startet das echte Programm per os/exec.
func ExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return nil, fmt.Errorf("%s: %s", name, strings.TrimSpace(string(ee.Stderr)))
	}
	return out, err
}

// Client kapselt kubectl config für genau eine Kubeconfig-Datei.
type Client struct {
	Kubeconfig string
	Run        Runner
}

func (c Client) config(ctx context.Context, args ...string) (string, error) {
	if c.Kubeconfig == "" {
		return "", errors.New("keine Kubeconfig angegeben")
	}
	if _, err := os.Stat(c.Kubeconfig); err != nil {
		return "", err
	}
	args = append([]string{"config"}, args...)
	args = append(args, "--kubeconfig", c.Kubeconfig)
	out, err := c.Run(ctx, "kubectl", args...)
	return strings.TrimSpace(string(out)), err
}

// Contexts listet alle Kontextnamen.
func (c Client) Contexts(ctx context.Context) ([]string, error) {
	out, err := c.config(ctx, "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// Current liefert den aktiven Kontext.
func (c Client) Current(ctx context.Context) (string, error) {
	return c.config(ctx, "current-context")
}

// Use setzt den aktiven Kontext.
func (c Client) Use(ctx context.Context, name string) error {
	_, err := c.config(ctx, "use-context", name)
	return err
}
