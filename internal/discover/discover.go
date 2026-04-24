package discover

import (
	"fmt"
	"os"
	"path/filepath"
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
func Params(dir string) ([]Parameter, error) {
	files, err := Files(dir)
	if err != nil {
		return nil, err
	}

	if scoreutil.IsSingleMode(dir) {
		return []Parameter{{Name: "image", Required: true, String: ""}}, nil
	}

	params := make([]Parameter, 0, len(files))
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
	return params, nil
}
