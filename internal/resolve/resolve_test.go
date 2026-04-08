package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const scoreYAML = `apiVersion: score.dev/v1b1
metadata:
  name: %s
containers:
  main:
    image: .
`

func TestRun_SingleMode_FromJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))

	t.Setenv("ARGOCD_APP_PARAMETERS", `[{"name":"image","string":"registry/myapp:v1.0"}]`)

	img, err := Run(filepath.Join(dir, "score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "registry/myapp:v1.0" {
		t.Errorf("got %q, want %q", img, "registry/myapp:v1.0")
	}
}

func TestRun_MultiMode_FromJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", fmt.Sprintf(scoreYAML, "backend"))
	writeFile(t, dir, "frontend.score.yaml", fmt.Sprintf(scoreYAML, "frontend"))

	t.Setenv("ARGOCD_APP_PARAMETERS", `[{"name":"image-backend","string":"reg/backend:v1"},{"name":"image-frontend","string":"reg/frontend:v2"}]`)

	img, err := Run(filepath.Join(dir, "backend.score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "reg/backend:v1" {
		t.Errorf("got %q, want %q", img, "reg/backend:v1")
	}

	img, err = Run(filepath.Join(dir, "frontend.score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "reg/frontend:v2" {
		t.Errorf("got %q, want %q", img, "reg/frontend:v2")
	}
}

func TestRun_FallbackToParamEnv(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", fmt.Sprintf(scoreYAML, "backend"))
	writeFile(t, dir, "frontend.score.yaml", fmt.Sprintf(scoreYAML, "frontend"))

	t.Setenv("ARGOCD_APP_PARAMETERS", "")
	t.Setenv("PARAM_IMAGE_BACKEND", "reg/backend:v3")

	img, err := Run(filepath.Join(dir, "backend.score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "reg/backend:v3" {
		t.Errorf("got %q, want %q", img, "reg/backend:v3")
	}
}

func TestRun_FallbackToPlaceholder(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))

	t.Setenv("ARGOCD_APP_PARAMETERS", "")
	t.Setenv("PARAM_IMAGE", "")

	img, err := Run(filepath.Join(dir, "score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "placeholder:latest" {
		t.Errorf("got %q, want %q", img, "placeholder:latest")
	}
}

func TestRun_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))

	t.Setenv("ARGOCD_APP_PARAMETERS", `not-json`)

	_, err := Run(filepath.Join(dir, "score.yaml"), dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRun_HyphenInName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "my-service.score.yaml", fmt.Sprintf(scoreYAML, "my-service"))
	writeFile(t, dir, "other.score.yaml", fmt.Sprintf(scoreYAML, "other"))

	t.Setenv("ARGOCD_APP_PARAMETERS", `[{"name":"image-my-service","string":"reg/svc:v1"}]`)

	img, err := Run(filepath.Join(dir, "my-service.score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "reg/svc:v1" {
		t.Errorf("got %q, want %q", img, "reg/svc:v1")
	}
}

func TestRun_HyphenFallbackToParamEnv(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "my-service.score.yaml", fmt.Sprintf(scoreYAML, "my-service"))
	writeFile(t, dir, "other.score.yaml", fmt.Sprintf(scoreYAML, "other"))

	t.Setenv("ARGOCD_APP_PARAMETERS", "")
	t.Setenv("PARAM_IMAGE_MY_SERVICE", "reg/svc:v2")

	img, err := Run(filepath.Join(dir, "my-service.score.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if img != "reg/svc:v2" {
		t.Errorf("got %q, want %q", img, "reg/svc:v2")
	}
}
