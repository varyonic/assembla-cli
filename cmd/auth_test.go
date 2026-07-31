package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStderr runs fn with os.Stderr redirected, and returns what it wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stderr")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}

	prev := os.Stderr
	os.Stderr = file
	func() {
		defer func() {
			os.Stderr = prev
			file.Close()
		}()
		fn()
	}()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestProjectSecretWarningInsideRepository: the message must say what was done
// and that it does not cover an already-tracked file.
func TestProjectSecretWarningInsideRepository(t *testing.T) {
	note := "added .assembla.yml to /repo/.git/info/exclude"

	got := captureStderr(t, func() { printProjectSecretWarning(note) })

	for _, want := range []string{
		"WARNING",
		"API key and secret",
		note,
		"already-tracked",
		"git rm --cached .assembla.yml",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("warning should mention %q; got:\n%s", want, got)
		}
	}
}

// TestProjectSecretWarningOutsideRepository must not claim a git rule was added.
func TestProjectSecretWarningOutsideRepository(t *testing.T) {
	got := captureStderr(t, func() { printProjectSecretWarning("") })

	if !strings.Contains(got, "WARNING") || !strings.Contains(got, "API key and secret") {
		t.Errorf("warning should still flag the secrets; got:\n%s", got)
	}
	for _, unwanted := range []string{"git rm", "exclude", "already-tracked"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("warning should not mention %q outside a repository; got:\n%s", unwanted, got)
		}
	}
}
