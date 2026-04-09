package discover

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

func TestFiles_Single(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))

	files, err := Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "score.yaml" {
		t.Errorf("got %v, want [score.yaml]", files)
	}
}

func TestFiles_Multi(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", fmt.Sprintf(scoreYAML, "backend"))
	writeFile(t, dir, "frontend.score.yaml", fmt.Sprintf(scoreYAML, "frontend"))

	files, err := Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if files[0] != "backend.score.yaml" || files[1] != "frontend.score.yaml" {
		t.Errorf("got %v", files)
	}
}

func TestFiles_ScoreYAMLTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))
	writeFile(t, dir, "extra.score.yaml", fmt.Sprintf(scoreYAML, "extra"))

	files, err := Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "score.yaml" {
		t.Errorf("got %v, want [score.yaml]", files)
	}
}

func TestFiles_NoFiles(t *testing.T) {
	dir := t.TempDir()
	_, err := Files(dir)
	if err == nil {
		t.Fatal("expected error for 0 score files")
	}
}

func TestParams_SingleMode(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))

	params, err := Params(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 1 {
		t.Fatalf("got %d params, want 1", len(params))
	}
	if params[0].Name != "image" {
		t.Errorf("got name %q, want %q", params[0].Name, "image")
	}
	if !params[0].Required {
		t.Error("expected required=true")
	}
}

func TestParams_MultiMode(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", fmt.Sprintf(scoreYAML, "backend"))
	writeFile(t, dir, "frontend.score.yaml", fmt.Sprintf(scoreYAML, "frontend"))

	params, err := Params(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 2 {
		t.Fatalf("got %d params, want 2", len(params))
	}
	if params[0].Name != "image-backend" {
		t.Errorf("got name %q, want %q", params[0].Name, "image-backend")
	}
	if params[1].Name != "image-frontend" {
		t.Errorf("got name %q, want %q", params[1].Name, "image-frontend")
	}
}

func TestParams_NoFiles(t *testing.T) {
	dir := t.TempDir()
	_, err := Params(dir)
	if err == nil {
		t.Fatal("expected error for 0 score files")
	}
}

func TestParams_ScoreYAMLTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", fmt.Sprintf(scoreYAML, "myapp"))
	writeFile(t, dir, "extra.score.yaml", fmt.Sprintf(scoreYAML, "extra"))

	params, err := Params(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 1 || params[0].Name != "image" {
		t.Errorf("got %v, want single image param", params)
	}
}
