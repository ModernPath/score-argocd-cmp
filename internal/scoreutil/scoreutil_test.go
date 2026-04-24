package scoreutil

import (
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

func TestParseWorkloadName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", `apiVersion: score.dev/v1b1
metadata:
  name: backend
containers:
  backend:
    image: .
`)

	name, err := ParseWorkloadName(filepath.Join(dir, "backend.score.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if name != "backend" {
		t.Errorf("got %q, want %q", name, "backend")
	}
}

func TestParseWorkloadName_MissingName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "bad.score.yaml", `apiVersion: score.dev/v1b1
metadata: {}
`)

	_, err := ParseWorkloadName(filepath.Join(dir, "bad.score.yaml"))
	if err == nil {
		t.Fatal("expected error for missing metadata.name")
	}
}

func TestParseWorkloadName_FileNotFound(t *testing.T) {
	_, err := ParseWorkloadName("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestListScoreFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", "")
	writeFile(t, dir, "frontend.score.yaml", "")
	writeFile(t, dir, "other.yaml", "")

	files, err := ListScoreFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
}

func TestListScoreFiles_Empty(t *testing.T) {
	dir := t.TempDir()
	files, err := ListScoreFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("got %d files, want 0", len(files))
	}
}

func TestIsSingleMode(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "score.yaml", "")
	if !IsSingleMode(dir) {
		t.Error("expected single mode when score.yaml exists")
	}
}

func TestIsSingleMode_Multi(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "backend.score.yaml", "")
	writeFile(t, dir, "frontend.score.yaml", "")
	if IsSingleMode(dir) {
		t.Error("expected multi mode when score.yaml does not exist")
	}
}

func TestIsSingleMode_Empty(t *testing.T) {
	dir := t.TempDir()
	if IsSingleMode(dir) {
		t.Error("expected non-single mode with 0 files")
	}
}
