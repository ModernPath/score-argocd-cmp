package paramscan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestScanDir_DiscoversAndDedupesAndExcludesBuiltins(t *testing.T) {
	dir := t.TempDir()

	// Inline Go template reference.
	writeFile(t, filepath.Join(dir, "dns.yaml"), `
- uri: cmd://sh#dns
  args:
    - "-c"
    - |
      echo '{"host":"{{ env "PARAM_DOMAIN" | default "localhost" }}"}'
`)

	// Built-in PARAM that must be filtered out.
	writeFile(t, filepath.Join(dir, "pg.yaml"), `
- uri: template://org/postgres
  init: |
    instanceName: {{ env "PARAM_INSTANCE_NAME" | default "local" }}
`)

	// Subdir, .yml extension, and JSON-style escaped quotes (simulating how
	// score-k8s might serialize provisioner template strings into its state).
	writeFile(t, filepath.Join(dir, "sub", "x.yml"), `
- uri: cmd://sh
  args:
    - "-c"
    - "echo {{ env \"PARAM_REGION\" }} {{ env \"PARAM_TIER\" }}"
`)

	// Shell-style substitution (used by cmd:// provisioner args, where Go
	// templates are not applied by score-k8s).
	writeFile(t, filepath.Join(dir, "shell.yaml"), `
- uri: cmd://sh
  args:
    - "-c"
    - 'echo "${PARAM_BUCKET:-default}"'
`)

	// Duplicate of PARAM_DOMAIN to verify dedupe.
	writeFile(t, filepath.Join(dir, "sub", "dup.yaml"), `# also uses {{ env "PARAM_DOMAIN" }}`)

	// Non-yaml file should be ignored even if it contains PARAM_*.
	writeFile(t, filepath.Join(dir, "readme.md"), "PARAM_IGNORED")

	got, err := ScanDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"PARAM_BUCKET", "PARAM_DOMAIN", "PARAM_REGION", "PARAM_TIER"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestScanDir_EmptyDir(t *testing.T) {
	got, err := ScanDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want []", got)
	}
}

func TestScanDir_MissingDirIsError(t *testing.T) {
	_, err := ScanDir(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected error scanning missing dir")
	}
}

func TestToParamName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"PARAM_DOMAIN", "domain"},
		{"PARAM_FOO_BAR", "foo-bar"},
		{"PARAM_X", "x"},
		{"PARAM_LONG_NAME_HERE", "long-name-here"},
	}
	for _, c := range cases {
		if got := ToParamName(c.in); got != c.want {
			t.Errorf("ToParamName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
