package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	ConfigDir        string
	GlobalConfigFile string
)

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
func LoadConfig() map[string]interface{} {
	config := make(map[string]interface{})

	// 1. Global config
	for k, v := range readYAML(GlobalConfigFile) {
		config[k] = v
	}

	// 2. Project config (overrides global)
	projectFile := findProjectConfig()
	if projectFile != "" {
		for k, v := range readYAML(projectFile) {
			config[k] = v
		}
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
		}
	}

	// Default API URL
	if _, ok := config["api_url"]; !ok {
		config["api_url"] = "https://api.assembla.com"
	}

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
