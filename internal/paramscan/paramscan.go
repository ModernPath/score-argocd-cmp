// Package paramscan finds plugin-parameter references inside provisioner
// definitions so that the CMP can announce them as ArgoCD dynamic parameters.
//
// The scan is intentionally permissive: it greps any PARAM_FOO_BAR identifier
// out of YAML files regardless of context (Go template, JSON-escaped string,
// shell variable, comment). The only place such tokens legitimately appear
// in a provisioner file is to reference a plugin parameter, so the rate of
// false positives is effectively zero in practice.
package paramscan

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// paramRegex matches PARAM_<UPPER_SNAKE> identifiers anywhere in a file.
var paramRegex = regexp.MustCompile(`PARAM_[A-Z0-9_]+`)

// builtIn lists plugin parameters that the CMP handles itself; these must
// never be re-announced as dynamic params (they are already declared as
// static params in plugin.yaml).
var builtIn = map[string]bool{
	"PARAM_PROVISIONERS_URL":        true,
	"PARAM_INSTANCE_NAME":           true,
	"PARAM_NO_DEFAULT_PROVISIONERS": true,
	"PARAM_DEBUG":                   true,
}

// ScanDir walks dir recursively, reads every *.yaml / *.yml file, and returns
// the unique sorted set of PARAM_* identifiers found, minus the built-ins.
func ScanDir(dir string) ([]string, error) {
	seen := map[string]struct{}{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".yaml", ".yml":
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range paramRegex.FindAll(data, -1) {
			name := string(m)
			if !builtIn[name] {
				seen[name] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

// ToParamName is the inverse of resolve.go's PARAM_ + UPPER(SNAKE) mapping.
// It converts PARAM_FOO_BAR -> "foo-bar", which is the form ArgoCD uses for
// plugin parameter names in spec.source.plugin.parameters[].name.
func ToParamName(envName string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(envName, "PARAM_"), "_", "-"))
}
