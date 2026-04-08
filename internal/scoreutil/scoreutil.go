package scoreutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type ScoreWorkload struct {
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

func ParseWorkloadName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	var w ScoreWorkload
	if err := yaml.Unmarshal(data, &w); err != nil {
		return "", fmt.Errorf("parsing %s: %w", path, err)
	}
	if w.Metadata.Name == "" {
		return "", fmt.Errorf("%s: metadata.name is empty", path)
	}
	return w.Metadata.Name, nil
}

func ListScoreFiles(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.score.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func IsSingleMode(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "score.yaml"))
	return err == nil
}
