package generate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"score-argocd-cmp/internal/debug"
	"score-argocd-cmp/internal/discover"
	"score-argocd-cmp/internal/resolve"
	"strings"
)

// Run discovers score files, resolves images, runs score-k8s generate for each,
// and returns the combined manifests. Returns an error if any step fails;
// never returns partial output.
func Run(dir string) (string, error) {
	files, err := discover.Files(dir)
	if err != nil {
		return "", fmt.Errorf("discover: %w", err)
	}
	debug.Logf("discovered score files: %v", files)

	namespace := os.Getenv("ARGOCD_APP_NAMESPACE")
	if namespace == "" {
		return "", fmt.Errorf("ARGOCD_APP_NAMESPACE is not set")
	}

	for _, f := range files {
		img, err := resolve.Run(filepath.Join(dir, f), dir)
		if err != nil {
			return "", fmt.Errorf("resolve-image %s: %w", f, err)
		}
		debug.Logf("resolved image for %s: %s", f, img)

		args := []string{}
		if debug.Enabled() {
			args = append(args, "--verbose")
		}
		args = append(args, "generate", f, "--image", img, "--namespace", namespace)
		debug.LogCmd("score-k8s", args)
		cmd := exec.Command("score-k8s", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("score-k8s generate %s: %w (output: %s)", f, err, string(out))
		}
	}

	manifestPath := filepath.Join(dir, "manifests.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("reading manifests.yaml: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("manifests.yaml is empty after generation")
	}

	return content, nil
}
