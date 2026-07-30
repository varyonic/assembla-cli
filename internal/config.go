package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ConfigDir        string
	GlobalConfigFile string
)

// Keys a configuration file is allowed to set. The default is deny: a key that
// some part of the code happens to read does not become file-settable unless it
// is listed here, and the loader's own "_" markers can never be forged.
var (
	globalConfigKeys = map[string]bool{
		"api_key":    true,
		"api_secret": true,
		"space":      true,
		"api_url":    true,
	}

	// A project .assembla.yml can arrive as part of a cloned repository, so it does
	// not get to choose where credentials are sent: api_url is deliberately absent.
	// Set it in the global config or ASSEMBLA_API_URL instead.
	projectConfigKeys = map[string]bool{
		"api_key":    true,
		"api_secret": true,
		"space":      true,
	}
)

// mergeAllowedKeys copies the permitted keys of src into dst, ignoring the rest.
func mergeAllowedKeys(dst, src map[string]interface{}, allowed map[string]bool) {
	for key, value := range src {
		if allowed[key] {
			dst[key] = value
		}
	}
}

// warnAboutIgnoredKeys reports keys the loader will not act on, so a misspelled
// setting does not fail silently. Keys are reported in a stable order.
func warnAboutIgnoredKeys(path string, fileConfig map[string]interface{}, allowed map[string]bool) {
	var ignored []string
	for key := range fileConfig {
		if !allowed[key] {
			ignored = append(ignored, key)
		}
	}
	sort.Strings(ignored)

	for _, key := range ignored {
		switch {
		case globalConfigKeys[key]:
			// Recognised, but deliberately not honoured from this file.
			fmt.Fprintf(os.Stderr, "Warning: %s: ignoring %q, which is only honoured in %s\n",
				path, key, GlobalConfigFile)
		case caseInsensitiveMatch(key, allowed) != "":
			fmt.Fprintf(os.Stderr, "Warning: %s: ignoring unrecognised key %q (did you mean %q?)\n",
				path, key, caseInsensitiveMatch(key, allowed))
		default:
			fmt.Fprintf(os.Stderr, "Warning: %s: ignoring unrecognised key %q\n", path, key)
		}
	}
}

// caseInsensitiveMatch returns the allowed key that differs from key only in
// case, which is the most common way a setting is misspelled.
func caseInsensitiveMatch(key string, allowed map[string]bool) string {
	for candidate := range allowed {
		if strings.EqualFold(candidate, key) {
			return candidate
		}
	}
	return ""
}

// suppliedAPIURL reports whether a config file both contains api_url and is
// permitted to set it, so provenance stays correct if the allowlist changes.
func suppliedAPIURL(fileConfig map[string]interface{}, allowed map[string]bool) bool {
	if !allowed["api_url"] {
		return false
	}
	_, ok := fileConfig["api_url"]
	return ok
}

func init() {
	home, _ := os.UserHomeDir()
	ConfigDir = filepath.Join(home, ".config", "assembla")
	GlobalConfigFile = filepath.Join(ConfigDir, "config.yml")
}

func findProjectConfig() string {
	current, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(current, ".assembla.yml")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func readYAML(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]interface{}{}
	}
	result := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

// LoadConfig loads configuration with precedence: env vars > project .assembla.yml > global config.
//
// The origin of api_url is recorded in _api_url_source, because credentials are
// sent to that host and project config is not necessarily trustworthy.
func LoadConfig() map[string]interface{} {
	config := make(map[string]interface{})
	apiURLSource := SourceDefault

	// 1. Global config
	globalConfig := readYAML(GlobalConfigFile)
	warnAboutIgnoredKeys(GlobalConfigFile, globalConfig, globalConfigKeys)
	mergeAllowedKeys(config, globalConfig, globalConfigKeys)
	if suppliedAPIURL(globalConfig, globalConfigKeys) {
		apiURLSource = SourceGlobal
	}

	// 2. Project config (overrides global, except for api_url which it may not set)
	projectFile := findProjectConfig()
	if projectFile != "" {
		projectConfig := readYAML(projectFile)
		warnAboutIgnoredKeys(projectFile, projectConfig, projectConfigKeys)
		mergeAllowedKeys(config, projectConfig, projectConfigKeys)
		config["_project_config"] = projectFile
	}

	// 3. Environment variables (highest precedence)
	envMap := map[string]string{
		"ASSEMBLA_API_KEY":    "api_key",
		"ASSEMBLA_API_SECRET": "api_secret",
		"ASSEMBLA_SPACE":      "space",
		"ASSEMBLA_API_URL":    "api_url",
	}
	for envVar, key := range envMap {
		if value := os.Getenv(envVar); value != "" {
			config[key] = value
			if key == "api_url" {
				apiURLSource = SourceEnv
			}
		}
	}

	// Default API URL
	if _, ok := config["api_url"]; !ok {
		config["api_url"] = DefaultAPIURL
	}
	config["_api_url_source"] = apiURLSource

	return config
}

// SaveGlobalConfig merges data into the global config file.
func SaveGlobalConfig(data map[string]interface{}) (string, error) {
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		return "", err
	}
	existing := readYAML(GlobalConfigFile)
	for k, v := range data {
		existing[k] = v
	}
	out, err := yaml.Marshal(existing)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(GlobalConfigFile, out, 0644); err != nil {
		return "", err
	}
	return GlobalConfigFile, nil
}

// SaveProjectConfig merges data into a project config file.
func SaveProjectConfig(data map[string]interface{}, path string) (string, error) {
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, ".assembla.yml")
	}
	existing := readYAML(path)
	for k, v := range data {
		existing[k] = v
	}
	out, err := yaml.Marshal(existing)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return "", err
	}
	return path, nil
}
