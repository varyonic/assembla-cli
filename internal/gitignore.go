package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfigName is the per-project config file name.
const ProjectConfigName = ".assembla.yml"

// ProjectConfigPath returns the project config path for the working directory.
func ProjectConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ProjectConfigName), nil
}

// EnsureProjectConfigIgnored keeps a project config that holds credentials out of
// version control.
//
// It writes to .git/info/exclude rather than .gitignore: the exclusion matters to
// this clone, and committing a rule to everyone else's .gitignore is not this
// tool's decision to make.
//
// The returned string describes what happened, for display to the user. It is
// empty when configPath is not inside a git repository.
func EnsureProjectConfigIgnored(configPath string) (string, error) {
	root := findRepoRoot(filepath.Dir(configPath))
	if root == "" {
		return "", nil
	}

	gitDir, err := resolveGitDir(root)
	if err != nil {
		return "", err
	}

	if where := existingIgnoreRule(root, gitDir); where != "" {
		return fmt.Sprintf("%s is already ignored via %s", ProjectConfigName, where), nil
	}

	exclude := filepath.Join(gitDir, "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(exclude), 0755); err != nil {
		return "", err
	}
	if err := appendLine(exclude, ProjectConfigName); err != nil {
		return "", err
	}
	return fmt.Sprintf("added %s to %s", ProjectConfigName, exclude), nil
}

// findRepoRoot walks up from dir looking for a git repository, returning its
// working-tree root or "" if there is none.
func findRepoRoot(dir string) string {
	for {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// resolveGitDir returns the repository's git directory, following the "gitdir:"
// pointer that linked worktrees and submodules use in place of a .git directory.
func resolveGitDir(root string) (string, error) {
	dotGit := filepath.Join(root, ".git")

	info, err := os.Stat(dotGit)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return dotGit, nil
	}

	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "gitdir:") {
			continue
		}
		target := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
		if target == "" {
			break
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(root, target)
		}
		return filepath.Clean(target), nil
	}
	return "", fmt.Errorf("could not read a gitdir path from %s", dotGit)
}

// existingIgnoreRule reports which file already ignores the project config, so a
// redundant rule is not appended. It only recognises a plain entry for the exact
// filename; a broader pattern elsewhere may already cover it, in which case the
// extra exclude line is harmless.
func existingIgnoreRule(root, gitDir string) string {
	for _, candidate := range []string{
		filepath.Join(root, ".gitignore"),
		filepath.Join(gitDir, "info", "exclude"),
	} {
		if hasIgnoreLine(candidate) {
			return candidate
		}
	}
	return ""
}

func hasIgnoreLine(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		switch strings.TrimSpace(scanner.Text()) {
		case ProjectConfigName, "/" + ProjectConfigName:
			return true
		}
	}
	return false
}

// appendLine adds line to path, creating the file and repairing a missing
// trailing newline so the new rule does not merge into the previous one.
func appendLine(path, line string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var out strings.Builder
	out.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		out.WriteString("\n")
	}
	out.WriteString(line + "\n")

	// Not a secret: exclude files are ordinary repository metadata.
	return os.WriteFile(path, []byte(out.String()), 0644)
}
