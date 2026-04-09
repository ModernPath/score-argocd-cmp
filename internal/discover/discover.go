package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"score-argocd-cmp/internal/debug"
	"score-argocd-cmp/internal/initialize"
	"score-argocd-cmp/internal/paramscan"
	"score-argocd-cmp/internal/scoreutil"
)

type Parameter struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	String   string `json:"string"`
}

// Files returns the list of score files in the directory.
// Returns score.yaml if it exists, otherwise all *.score.yaml files.
// Returns error if no score files are found.
func Files(dir string) ([]string, error) {
	single := filepath.Join(dir, "score.yaml")
	if _, err := os.Stat(single); err == nil {
		return []string{"score.yaml"}, nil
	}

	files, err := scoreutil.ListScoreFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no score.yaml or *.score.yaml files found in %s", dir)
	}

	// Return basenames only
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = filepath.Base(f)
	}
	return names, nil
}

// Params returns the parameter announcements for ArgoCD dynamic parameters.
// Image params are derived from the score files in dir; additional params
// are auto-discovered by scanning the loaded provisioners for PARAM_*
// references (best-effort — failures here do not fail the RPC).
func Params(dir string) ([]Parameter, error) {
	files, err := Files(dir)
	if err != nil {
		return nil, err
	}

	var params []Parameter
	if scoreutil.IsSingleMode(dir) {
		params = []Parameter{{Name: "image", Required: true, String: ""}}
	} else {
		params = make([]Parameter, 0, len(files))
		for _, f := range files {
			name, err := scoreutil.ParseWorkloadName(filepath.Join(dir, f))
			if err != nil {
				return nil, err
			}
			params = append(params, Parameter{
				Name:     "image-" + name,
				Required: true,
				String:   "",
			})
		}
	}

	extra, err := discoverProvisionerParams()
	if err != nil {
		// Don't fail the whole RPC just because the OCI registry was
		// unreachable or PARAM_PROVISIONERS_URL was unset — the static
		// image params alone are still useful.
		debug.Logf("provisioner param discovery failed: %v", err)
	} else {
		seen := map[string]bool{}
		for _, p := range params {
			seen[p.Name] = true
		}
		for _, name := range extra {
			if seen[name] {
				continue
			}
			params = append(params, Parameter{Name: name})
			seen[name] = true
		}
	}

	return params, nil
}

// discoverProvisionerParams runs score-k8s init in a temporary directory to
// materialize whatever provisioners PARAM_PROVISIONERS_URL points at, then
// scans the resulting state for PARAM_* references and converts them to
// ArgoCD parameter names (e.g. PARAM_DOMAIN -> "domain"). Returns (nil, nil)
// if PARAM_PROVISIONERS_URL is unset, so callers can use this safely outside
// of a real ArgoCD CMP environment.
func discoverProvisionerParams() ([]string, error) {
	if os.Getenv("PARAM_PROVISIONERS_URL") == "" {
		return nil, nil
	}
	tmp, err := os.MkdirTemp("", "score-cmp-discover-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	if err := initialize.RunInDir(tmp, nil); err != nil {
		return nil, fmt.Errorf("init in temp dir: %w", err)
	}
	envNames, err := paramscan.ScanDir(filepath.Join(tmp, ".score-k8s"))
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	out := make([]string, 0, len(envNames))
	for _, n := range envNames {
		out = append(out, paramscan.ToParamName(n))
	}
	return out, nil
}
