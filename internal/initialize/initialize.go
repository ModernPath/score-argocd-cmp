package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"score-argocd-cmp/internal/debug"
)

// Run executes score-k8s init with the appropriate provisioners flags
// derived from PARAM_PROVISIONERS_URL env var.
// Any extra arguments are passed through to score-k8s init.
func Run(extra []string) error {
	args := []string{"init", "--no-sample"}

	url := os.Getenv("PARAM_PROVISIONERS_URL")
	if url == "" {
		return fmt.Errorf("PARAM_PROVISIONERS_URL is not set")
	}
	args = append(args, "--provisioners", url)

	args = append(args, extra...)

	debug.LogCmd("score-k8s", args)
	cmd := exec.Command("score-k8s", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("score-k8s init: %w (output: %s)", err, string(out))
	}
	return nil
}
