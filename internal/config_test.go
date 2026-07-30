package internal

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadConfigAPIURLProvenance: diagnostics name the source of a rejected
// api_url, so the recorded origin has to be right. Project config never appears
// here — it is not allowed to set api_url at all.
func TestLoadConfigAPIURLProvenance(t *testing.T) {
	tests := []struct {
		name       string
		global     string
		project    string
		envURL     string
		wantSource string
		wantURL    string
	}{
		{
			name:       "nothing configured",
			wantSource: SourceDefault,
			wantURL:    DefaultAPIURL,
		},
		{
			name:       "global only",
			global:     "api_url: https://global.example.com\n",
			wantSource: SourceGlobal,
			wantURL:    "https://global.example.com",
		},
		{
			name:       "project cannot set it, so the default stands",
			project:    "api_url: https://project.example.com\n",
			wantSource: SourceDefault,
			wantURL:    DefaultAPIURL,
		},
		{
			name:       "project cannot override global",
			global:     "api_url: https://global.example.com\n",
			project:    "api_url: https://project.example.com\n",
			wantSource: SourceGlobal,
			wantURL:    "https://global.example.com",
		},
		{
			name:       "env overrides global",
			global:     "api_url: https://global.example.com\n",
			envURL:     "https://env.example.com",
			wantSource: SourceEnv,
			wantURL:    "https://env.example.com",
		},
		{
			name:       "env wins even when a project tries",
			project:    "api_url: https://project.example.com\n",
			envURL:     "https://env.example.com",
			wantSource: SourceEnv,
			wantURL:    "https://env.example.com",
		},
		{
			name:       "project config with other keys leaves global source",
			global:     "api_url: https://global.example.com\n",
			project:    "space: myspace\n",
			wantSource: SourceGlobal,
			wantURL:    "https://global.example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearAssemblaEnv(t)
			configFile := withTempGlobalConfig(t)
			if tc.global != "" {
				writeFile(t, configFile, tc.global)
			}

			work := isolatedDir(t)
			if tc.project != "" {
				writeFile(t, filepath.Join(work, ".assembla.yml"), tc.project)
			}
			chdir(t, work)

			if tc.envURL != "" {
				t.Setenv("ASSEMBLA_API_URL", tc.envURL)
			}

			config := LoadConfig()

			if got := config["_api_url_source"]; got != tc.wantSource {
				t.Errorf("_api_url_source = %v, want %v", got, tc.wantSource)
			}
			if got := config["api_url"]; got != tc.wantURL {
				t.Errorf("api_url = %v, want %v", got, tc.wantURL)
			}
		})
	}
}

// TestProjectConfigCannotRedirectCredentials is the central guarantee, and it now
// holds structurally: the value never reaches the config, so nothing downstream
// has to defend against it.
func TestProjectConfigCannotRedirectCredentials(t *testing.T) {
	clearAssemblaEnv(t)
	configFile := withTempGlobalConfig(t)
	writeFile(t, configFile, "api_key: KEY\napi_secret: SECRET\n")

	work := isolatedDir(t)
	writeFile(t, filepath.Join(work, ".assembla.yml"),
		"api_url: https://evil.example.com\nspace: myspace\n")
	chdir(t, work)

	config := LoadConfig()

	if got := config["api_url"]; got != DefaultAPIURL {
		t.Errorf("api_url = %v, want the default despite the project config", got)
	}

	// The rest of the file is still honoured.
	if got := config["space"]; got != "myspace" {
		t.Errorf("space = %v, want myspace", got)
	}

	// And the value that does reach the client is the safe one.
	rawURL, _ := config["api_url"].(string)
	source, _ := config["_api_url_source"].(string)
	resolved, err := ResolveAPIURL(rawURL, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != DefaultAPIURL {
		t.Errorf("resolved %q, want %q", resolved, DefaultAPIURL)
	}
}

// TestLoadConfigFindsProjectConfigInParentDirectory documents the upward walk,
// which is why a project config is treated as untrusted input.
func TestLoadConfigFindsProjectConfigInParentDirectory(t *testing.T) {
	clearAssemblaEnv(t)
	withTempGlobalConfig(t)

	root := isolatedDir(t)
	nested := filepath.Join(root, "a", "b", "c")
	writeFile(t, filepath.Join(nested, ".keep"), "")
	writeFile(t, filepath.Join(root, ".assembla.yml"), "space: parent-space\n")
	chdir(t, nested)

	config := LoadConfig()

	if got := config["space"]; got != "parent-space" {
		t.Errorf("space = %v, want the parent directory's value", got)
	}
	if got, _ := config["_project_config"].(string); !strings.HasSuffix(got, ".assembla.yml") {
		t.Errorf("_project_config = %q, want the discovered file path", got)
	}
}

// TestSaveGlobalConfigMergesRatherThanReplaces protects credentials from being
// dropped when another setting is written.
func TestSaveGlobalConfigMergesRatherThanReplaces(t *testing.T) {
	configFile := withTempGlobalConfig(t)
	writeFile(t, configFile, "api_key: KEY\napi_secret: SECRET\nspace: myspace\n")

	if _, err := SaveGlobalConfig(map[string]interface{}{"api_url": "https://onprem.example.com"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got := readFileOrEmpty(t, configFile)
	for _, want := range []string{"KEY", "SECRET", "myspace", "onprem.example.com"} {
		if !strings.Contains(got, want) {
			t.Errorf("config lost %q; got:\n%s", want, got)
		}
	}
}
