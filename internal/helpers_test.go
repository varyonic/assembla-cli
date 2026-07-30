package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// These helpers redirect every piece of global state the package touches at a
// temporary location, so tests never read or write the real ~/.config/assembla.

// withTempGlobalConfig points ConfigDir/GlobalConfigFile at a temp directory and
// returns the config file path.
func withTempGlobalConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	prevDir, prevFile := ConfigDir, GlobalConfigFile
	ConfigDir = dir
	GlobalConfigFile = filepath.Join(dir, "config.yml")
	t.Cleanup(func() {
		ConfigDir, GlobalConfigFile = prevDir, prevFile
	})
	return GlobalConfigFile
}

// writeFile writes content to path, creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

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

	return readFileOrEmpty(t, path)
}

// readFileOrEmpty returns a file's contents, or "" if it does not exist.
func readFileOrEmpty(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// chdir switches to dir for the duration of the test. Tests using it must not
// call t.Parallel, since the working directory is process-wide.
func chdir(t *testing.T, dir string) {
	t.Helper()

	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

// isolatedDir returns a temp directory with a .assembla.yml sentinel above it,
// so findProjectConfig cannot wander into a real config outside the test.
func isolatedDir(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	dir := filepath.Join(root, "work")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
}

// clearAssemblaEnv unsets the ASSEMBLA_* variables so ambient environment does
// not leak into config precedence tests.
func clearAssemblaEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"ASSEMBLA_API_KEY", "ASSEMBLA_API_SECRET", "ASSEMBLA_SPACE", "ASSEMBLA_API_URL",
	} {
		t.Setenv(key, "")
	}
}
