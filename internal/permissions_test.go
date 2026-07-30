package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertOwnerOnly fails if anyone other than the owner can read path.
func assertOwnerOnly(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if perm := info.Mode().Perm(); perm&0077 != 0 {
		t.Errorf("%s is mode %04o, want no group/other access", path, perm)
	}
}

func TestSaveGlobalConfigCreatesOwnerOnlyFileAndDir(t *testing.T) {
	configFile := withTempGlobalConfig(t)
	// Start from a directory that does not exist yet, as on a first login.
	ConfigDir = filepath.Join(ConfigDir, "assembla")
	GlobalConfigFile = filepath.Join(ConfigDir, "config.yml")
	configFile = GlobalConfigFile

	if _, err := SaveGlobalConfig(map[string]interface{}{
		"api_key":    "KEY",
		"api_secret": "SECRET",
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	assertOwnerOnly(t, configFile)
	assertOwnerOnly(t, ConfigDir)
}

// TestSaveGlobalConfigTightensExistingFile is the migration case: os.WriteFile
// leaves an existing file's mode alone, so a 0644 config from an earlier version
// would otherwise stay world-readable forever.
func TestSaveGlobalConfigTightensExistingFile(t *testing.T) {
	configFile := withTempGlobalConfig(t)
	if err := os.WriteFile(configFile, []byte("api_key: OLD\n"), 0644); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	if err := os.Chmod(configFile, 0644); err != nil {
		t.Fatalf("chmod seed: %v", err)
	}

	if _, err := SaveGlobalConfig(map[string]interface{}{"api_secret": "SECRET"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	assertOwnerOnly(t, configFile)
	if got := readFileOrEmpty(t, configFile); !strings.Contains(got, "SECRET") {
		t.Errorf("config should contain the new secret, got:\n%s", got)
	}
}

func TestSaveGlobalConfigTightensExistingDirectory(t *testing.T) {
	withTempGlobalConfig(t)
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(ConfigDir, 0755); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	if _, err := SaveGlobalConfig(map[string]interface{}{"api_key": "KEY"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	assertOwnerOnly(t, ConfigDir)
}

func TestSaveProjectConfigCreatesOwnerOnlyFile(t *testing.T) {
	dir := isolatedDir(t)
	path := filepath.Join(dir, ".assembla.yml")

	if _, err := SaveProjectConfig(map[string]interface{}{
		"api_key":    "KEY",
		"api_secret": "SECRET",
	}, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	assertOwnerOnly(t, path)
}

func TestSaveProjectConfigTightensExistingFile(t *testing.T) {
	dir := isolatedDir(t)
	path := filepath.Join(dir, ".assembla.yml")
	if err := os.WriteFile(path, []byte("space: myspace\n"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatalf("chmod seed: %v", err)
	}

	if _, err := SaveProjectConfig(map[string]interface{}{"api_secret": "SECRET"}, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	assertOwnerOnly(t, path)
}

// TestSaveProjectConfigDefaultsToWorkingDirectory covers the path used by
// `auth login --scope project`, which passes an empty path.
func TestSaveProjectConfigDefaultsToWorkingDirectory(t *testing.T) {
	dir := isolatedDir(t)
	chdir(t, dir)

	path, err := SaveProjectConfig(map[string]interface{}{"api_secret": "SECRET"}, "")
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if filepath.Base(path) != ".assembla.yml" {
		t.Errorf("path = %q, want .assembla.yml in the working directory", path)
	}
	assertOwnerOnly(t, path)
}

// TestWriteSecretFileFollowsSymlinks protects a config symlinked into a dotfiles
// repository: the link must survive and its target must be tightened.
func TestWriteSecretFileFollowsSymlinks(t *testing.T) {
	dir := isolatedDir(t)
	target := filepath.Join(dir, "real-config.yml")
	link := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(target, []byte("api_key: OLD\n"), 0644); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := os.Chmod(target, 0644); err != nil {
		t.Fatalf("chmod target: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := writeSecretFile(link, []byte("api_key: NEW\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced by a regular file")
	}
	if got := readFileOrEmpty(t, target); !strings.Contains(got, "NEW") {
		t.Errorf("target not updated through the symlink, got: %q", got)
	}
	assertOwnerOnly(t, target)
}

func TestWriteSecretFileCreatesMissingFile(t *testing.T) {
	path := filepath.Join(isolatedDir(t), "new.yml")

	if err := writeSecretFile(path, []byte("api_key: KEY\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	assertOwnerOnly(t, path)
}

// TestLoadConfigWarnsAboutOpenPermissions covers the nudge for configs that
// already exist at 0644 and are never rewritten.
func TestLoadConfigWarnsAboutOpenPermissions(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		wantWarn bool
	}{
		{"world readable", 0644, true},
		{"group readable", 0640, true},
		{"world writable", 0666, true},
		{"owner only", 0600, false},
		{"owner read only", 0400, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearAssemblaEnv(t)
			configFile := withTempGlobalConfig(t)
			if err := os.WriteFile(configFile, []byte("api_key: KEY\n"), tc.mode); err != nil {
				t.Fatalf("seed: %v", err)
			}
			if err := os.Chmod(configFile, tc.mode); err != nil {
				t.Fatalf("chmod: %v", err)
			}
			chdir(t, isolatedDir(t))

			stderr := captureStderr(t, func() { LoadConfig() })

			if warned := strings.Contains(stderr, "Warning:"); warned != tc.wantWarn {
				t.Errorf("warned=%v, want %v (stderr: %q)", warned, tc.wantWarn, stderr)
			}
			if tc.wantWarn {
				for _, want := range []string{configFile, "chmod 0600"} {
					if !strings.Contains(stderr, want) {
						t.Errorf("warning should mention %q, got: %q", want, stderr)
					}
				}
			}
		})
	}
}

// TestLoadConfigSilentWhenNoConfigExists: a missing file is normal, not a warning.
func TestLoadConfigSilentWhenNoConfigExists(t *testing.T) {
	clearAssemblaEnv(t)
	withTempGlobalConfig(t)
	chdir(t, isolatedDir(t))

	if stderr := captureStderr(t, func() { LoadConfig() }); stderr != "" {
		t.Errorf("expected no output, got %q", stderr)
	}
}
