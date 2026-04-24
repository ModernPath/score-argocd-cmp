package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"score-argocd-cmp/internal/scoreutil"
	"strings"
)

type appParam struct {
	Name   string `json:"name"`
	String string `json:"string"`
}

func Run(scoreFile string, dir string) (string, error) {
	var targetName string

	if scoreutil.IsSingleMode(dir) {
		targetName = "image"
	} else {
		name, err := scoreutil.ParseWorkloadName(scoreFile)
		if err != nil {
			return "", err
		}
		targetName = "image-" + name
	}

	image, err := lookupParam(targetName)
	if err != nil {
		return "", err
	}
	if image == "" {
		return "placeholder:latest", nil
	}
	return image, nil
}

func lookupParam(name string) (string, error) {
	raw := os.Getenv("ARGOCD_APP_PARAMETERS")
	if raw == "" {
		// Fall back to PARAM_ env var
		envName := "PARAM_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		return os.Getenv(envName), nil
	}

	var params []appParam
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return "", fmt.Errorf("parsing ARGOCD_APP_PARAMETERS: %w", err)
	}

	for _, p := range params {
		if p.Name == name {
			return p.String, nil
		}
	}
	return "", nil
}
