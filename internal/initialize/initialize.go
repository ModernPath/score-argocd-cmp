package initialize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"score-argocd-cmp/internal/debug"
	"sort"
)

// patchTemplatesDir is the in-image directory holding patch templates that are
// applied to every render. The templates work around score-k8s implementation
// details (notably the random InstanceSuffix on workload labels) and are
// generic, so they ship with the CMP rather than per-customer provisioners.
// The Dockerfile copies patches/ from the repo to this path.
const patchTemplatesDir = "/usr/local/share/score-argocd-cmp/patches"

// Run executes score-k8s init with the appropriate provisioners flags
// derived from PARAM_PROVISIONERS_URL env var.
// Any extra arguments are passed through to score-k8s init.
func Run(extra []string) error {
	args := []string{}
	if debug.Enabled() {
		args = append(args, "--verbose")
	}
	args = append(args, "init", "--no-sample")

	url := os.Getenv("PARAM_PROVISIONERS_URL")
	if url == "" {
		return fmt.Errorf("PARAM_PROVISIONERS_URL is not set")
	}
	args = append(args, "--provisioners", url)

	if os.Getenv("PARAM_NO_DEFAULT_PROVISIONERS") == "true" {
		args = append(args, "--no-default-provisioners")
	}

	patchTemplates, err := discoverPatchTemplates(patchTemplatesDir)
	if err != nil {
		return fmt.Errorf("discover patch templates: %w", err)
	}
	for _, p := range patchTemplates {
		args = append(args, "--patch-templates", p)
	}

	args = append(args, extra...)

	debug.LogCmd("score-k8s", args)
	cmd := exec.Command("score-k8s", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("score-k8s init: %w (output: %s)", err, string(out))
	}
	return nil
}

// discoverPatchTemplates returns the absolute paths of all *.tpl files in dir,
// sorted lexicographically so that load order is deterministic. A missing
// directory is not an error (useful for local dev where the binary is run
// outside the container image).
func discoverPatchTemplates(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".tpl" {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}
