package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a minimal git repository layout and returns its root.
func initRepo(t *testing.T) string {
	t.Helper()

	root := isolatedDir(t)
	if err := os.MkdirAll(filepath.Join(root, ".git", "info"), 0755); err != nil {
		t.Fatalf("mkdir .git/info: %v", err)
	}
	return root
}

func excludeContents(t *testing.T, root string) string {
	t.Helper()
	return readFileOrEmpty(t, filepath.Join(root, ".git", "info", "exclude"))
}

func TestEnsureProjectConfigIgnoredAddsExcludeRule(t *testing.T) {
	root := initRepo(t)

	note, err := EnsureProjectConfigIgnored(filepath.Join(root, ProjectConfigName))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(note, "added") || !strings.Contains(note, ProjectConfigName) {
		t.Errorf("note = %q, want it to report the added rule", note)
	}
	if got := excludeContents(t, root); !strings.Contains(got, ProjectConfigName) {
		t.Errorf("exclude file = %q, want it to contain %q", got, ProjectConfigName)
	}
}

// TestEnsureProjectConfigIgnoredWorksFromSubdirectory: the config may sit in a
// subdirectory of the repository, not only at its root.
func TestEnsureProjectConfigIgnoredWorksFromSubdirectory(t *testing.T) {
	root := initRepo(t)
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if _, err := EnsureProjectConfigIgnored(filepath.Join(nested, ProjectConfigName)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := excludeContents(t, root); !strings.Contains(got, ProjectConfigName) {
		t.Errorf("exclude file = %q, want the rule at the repository root", got)
	}
}

func TestEnsureProjectConfigIgnoredOutsideRepositoryDoesNothing(t *testing.T) {
	dir := isolatedDir(t)

	note, err := EnsureProjectConfigIgnored(filepath.Join(dir, ProjectConfigName))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note != "" {
		t.Errorf("note = %q, want empty outside a repository", note)
	}
}

// TestEnsureProjectConfigIgnoredIsIdempotent guards against the rule piling up on
// every login.
func TestEnsureProjectConfigIgnoredIsIdempotent(t *testing.T) {
	root := initRepo(t)
	configPath := filepath.Join(root, ProjectConfigName)

	for i := 0; i < 3; i++ {
		if _, err := EnsureProjectConfigIgnored(configPath); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}

	if got := strings.Count(excludeContents(t, root), ProjectConfigName); got != 1 {
		t.Errorf("rule appears %d times, want exactly 1:\n%s", got, excludeContents(t, root))
	}
}

func TestEnsureProjectConfigIgnoredRecognisesExistingRules(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		content string
	}{
		{"gitignore plain", ".gitignore", ProjectConfigName + "\n"},
		{"gitignore rooted", ".gitignore", "/" + ProjectConfigName + "\n"},
		{"gitignore among others", ".gitignore", "build/\n" + ProjectConfigName + "\n*.log\n"},
		{"gitignore with trailing spaces", ".gitignore", "  " + ProjectConfigName + "  \n"},
		{"exclude file", filepath.Join(".git", "info", "exclude"), ProjectConfigName + "\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := initRepo(t)
			writeFile(t, filepath.Join(root, tc.file), tc.content)

			note, err := EnsureProjectConfigIgnored(filepath.Join(root, ProjectConfigName))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(note, "already ignored") {
				t.Errorf("note = %q, want it to report an existing rule", note)
			}
			if got := strings.Count(excludeContents(t, root), ProjectConfigName); got > 1 {
				t.Errorf("rule duplicated in exclude file:\n%s", excludeContents(t, root))
			}
		})
	}
}

// TestEnsureProjectConfigIgnoredPreservesExistingExcludes: appending must not
// damage rules that are already there.
func TestEnsureProjectConfigIgnoredPreservesExistingExcludes(t *testing.T) {
	root := initRepo(t)
	exclude := filepath.Join(root, ".git", "info", "exclude")
	writeFile(t, exclude, "# local excludes\nscratch/\n*.tmp\n")

	if _, err := EnsureProjectConfigIgnored(filepath.Join(root, ProjectConfigName)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := excludeContents(t, root)
	for _, want := range []string{"# local excludes", "scratch/", "*.tmp", ProjectConfigName} {
		if !strings.Contains(got, want) {
			t.Errorf("exclude file lost %q; got:\n%s", want, got)
		}
	}
}

// TestAppendLineRepairsMissingTrailingNewline: without this the new rule would
// merge into the previous one and silently ignore the wrong thing.
func TestAppendLineRepairsMissingTrailingNewline(t *testing.T) {
	root := initRepo(t)
	exclude := filepath.Join(root, ".git", "info", "exclude")
	writeFile(t, exclude, "scratch/")

	if _, err := EnsureProjectConfigIgnored(filepath.Join(root, ProjectConfigName)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := excludeContents(t, root)
	if strings.Contains(got, "scratch/"+ProjectConfigName) {
		t.Errorf("rule merged into the previous line:\n%s", got)
	}
	for _, want := range []string{"scratch/\n", ProjectConfigName + "\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("exclude file missing %q; got:\n%s", want, got)
		}
	}
}

// TestEnsureProjectConfigIgnoredInLinkedWorktree covers the .git-as-a-file layout
// used by `git worktree` and submodules.
func TestEnsureProjectConfigIgnoredInLinkedWorktree(t *testing.T) {
	base := isolatedDir(t)
	realGitDir := filepath.Join(base, "actual-git-dir")
	if err := os.MkdirAll(realGitDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	worktree := filepath.Join(base, "worktree")
	if err := os.MkdirAll(worktree, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(worktree, ".git"), "gitdir: "+realGitDir+"\n")

	if _, err := EnsureProjectConfigIgnored(filepath.Join(worktree, ProjectConfigName)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFileOrEmpty(t, filepath.Join(realGitDir, "info", "exclude"))
	if !strings.Contains(got, ProjectConfigName) {
		t.Errorf("rule should land in the linked git dir; got %q", got)
	}
}

func TestEnsureProjectConfigIgnoredWithRelativeGitdirPointer(t *testing.T) {
	base := isolatedDir(t)
	worktree := filepath.Join(base, "worktree")
	if err := os.MkdirAll(filepath.Join(worktree, "nested-git"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(worktree, ".git"), "gitdir: nested-git\n")

	if _, err := EnsureProjectConfigIgnored(filepath.Join(worktree, ProjectConfigName)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readFileOrEmpty(t, filepath.Join(worktree, "nested-git", "info", "exclude"))
	if !strings.Contains(got, ProjectConfigName) {
		t.Errorf("relative gitdir pointer not resolved; got %q", got)
	}
}

func TestEnsureProjectConfigIgnoredRejectsUnreadableGitFile(t *testing.T) {
	dir := isolatedDir(t)
	writeFile(t, filepath.Join(dir, ".git"), "not a gitdir pointer\n")

	if _, err := EnsureProjectConfigIgnored(filepath.Join(dir, ProjectConfigName)); err == nil {
		t.Error("expected an error for a .git file without a gitdir pointer")
	}
}

func TestProjectConfigPathUsesWorkingDirectory(t *testing.T) {
	dir := isolatedDir(t)
	chdir(t, dir)

	path, err := ProjectConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != ProjectConfigName {
		t.Errorf("path = %q, want it to end in %s", path, ProjectConfigName)
	}
	// The temp directory may be reached through a symlink (/var vs /private/var on
	// macOS), so compare resolved paths.
	wantDir, _ := filepath.EvalSymlinks(dir)
	gotDir, _ := filepath.EvalSymlinks(filepath.Dir(path))
	if gotDir != wantDir {
		t.Errorf("directory = %q, want %q", gotDir, wantDir)
	}
}
