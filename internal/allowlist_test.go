package internal

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestAllowlistedKeysStillLoad is the regression guard: inverting the default to
// deny must not stop any documented key from working.
func TestAllowlistedKeysStillLoad(t *testing.T) {
	for _, scope := range []string{"global", "project"} {
		t.Run(scope, func(t *testing.T) {
			clearAssemblaEnv(t)
			configFile := withTempGlobalConfig(t)
			work := isolatedDir(t)

			content := "api_key: KEY\napi_secret: SECRET\nspace: myspace\n"
			want := map[string]string{
				"api_key":    "KEY",
				"api_secret": "SECRET",
				"space":      "myspace",
			}
			if scope == "global" {
				// api_url is global-only, so it belongs in this case alone.
				content += "api_url: https://example.com\n"
				want["api_url"] = "https://example.com"
				writeFile(t, configFile, content)
			} else {
				writeFile(t, filepath.Join(work, ".assembla.yml"), content)
			}
			chdir(t, work)

			config := LoadConfig()
			for key, want := range want {
				if got := config[key]; got != want {
					t.Errorf("config[%q] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

// TestUnknownKeysAreIgnored: the point of the allowlist. A key the code does not
// declare as file-settable must not reach the config map at all.
func TestUnknownKeysAreIgnored(t *testing.T) {
	unknown := "" +
		"api_key: KEY\n" +
		"unexpected_key: value\n" +
		"API_KEY: WRONGCASE\n" +
		"api_urls: https://typo.example.com\n" +
		"debug: true\n"

	for _, scope := range []string{"global", "project"} {
		t.Run(scope, func(t *testing.T) {
			clearAssemblaEnv(t)
			configFile := withTempGlobalConfig(t)
			work := isolatedDir(t)

			if scope == "global" {
				writeFile(t, configFile, unknown)
			} else {
				writeFile(t, filepath.Join(work, ".assembla.yml"), unknown)
			}
			chdir(t, work)

			config := LoadConfig()

			if config["api_key"] != "KEY" {
				t.Errorf("declared key should still load, got %v", config["api_key"])
			}
			for _, key := range []string{"unexpected_key", "API_KEY", "api_urls", "debug"} {
				if value, ok := config[key]; ok {
					t.Errorf("undeclared key %q reached the config as %v", key, value)
				}
			}
		})
	}
}

// TestLoaderMarkersCannotBeForged: previously these survived only because the
// loader happened to overwrite them after merging. Now they are structurally
// unreachable from a file.
func TestLoaderMarkersCannotBeForged(t *testing.T) {
	clearAssemblaEnv(t)
	withTempGlobalConfig(t)

	work := isolatedDir(t)
	configPath := filepath.Join(work, ".assembla.yml")
	writeFile(t, configPath,
		"api_url: https://evil.example.com\n"+
			"_api_url_source: "+SourceEnv+"\n"+
			"_project_config: \"\"\n")
	chdir(t, work)

	config := LoadConfig()

	// The forged provenance must not survive: nothing in the file was honoured, so
	// the URL is still the built-in default.
	if got := config["_api_url_source"]; got != SourceDefault {
		t.Errorf("_api_url_source = %v, want %v", got, SourceDefault)
	}
	// t.TempDir may be reached through a symlink (/var vs /private/var), so compare
	// the discovered file rather than the literal path.
	if got, _ := config["_project_config"].(string); filepath.Base(got) != filepath.Base(configPath) {
		t.Errorf("_project_config = %q, want the real project file", got)
	}
	if got := config["api_url"]; got != DefaultAPIURL {
		t.Errorf("api_url = %v, want the default", got)
	}
}

// TestWarnsAboutIgnoredKeys: an ignored key must be reported, or a misspelled
// setting looks like it took effect.
func TestWarnsAboutIgnoredKeys(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantWarn []string
		noWarn   []string
	}{
		{
			name:     "unrecognised key",
			content:  "api_key: KEY\ndebug: true\n",
			wantWarn: []string{"unrecognised key", `"debug"`},
		},
		{
			name:     "case typo suggests the real key",
			content:  "API_KEY: KEY\n",
			wantWarn: []string{`"API_KEY"`, "did you mean", `"api_key"`},
		},
		{
			name:     "near miss without a case match gets no suggestion",
			content:  "api_urls: https://example.com\n",
			wantWarn: []string{`"api_urls"`, "unrecognised key"},
			noWarn:   []string{"did you mean"},
		},
		{
			name:    "every key recognised",
			content: "api_key: K\napi_secret: S\nspace: sp\napi_url: https://example.com\n",
			noWarn:  []string{"Warning"},
		},
		{
			name:    "empty file",
			content: "",
			noWarn:  []string{"Warning"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearAssemblaEnv(t)
			configFile := withTempGlobalConfig(t)
			writeFile(t, configFile, tc.content)
			chdir(t, isolatedDir(t))

			stderr := captureStderr(t, func() { LoadConfig() })

			for _, want := range tc.wantWarn {
				if !strings.Contains(stderr, want) {
					t.Errorf("warning should mention %q; got:\n%s", want, stderr)
				}
			}
			for _, unwanted := range tc.noWarn {
				if strings.Contains(stderr, unwanted) {
					t.Errorf("output should not mention %q; got:\n%s", unwanted, stderr)
				}
			}
		})
	}
}

// TestWarnsThatAPIURLIsGlobalOnly: api_url is recognised, so calling it
// "unrecognised" would send the user hunting for a typo that is not there. This
// warning is now the visible signal that a repository tried to redirect the API.
func TestWarnsThatAPIURLIsGlobalOnly(t *testing.T) {
	clearAssemblaEnv(t)
	withTempGlobalConfig(t)

	work := isolatedDir(t)
	writeFile(t, filepath.Join(work, ".assembla.yml"),
		"api_url: https://evil.example.com\n")
	chdir(t, work)

	stderr := captureStderr(t, func() { LoadConfig() })

	if !strings.Contains(stderr, "only honoured in") {
		t.Errorf("warning should explain where the key is honoured; got:\n%s", stderr)
	}
	if strings.Contains(stderr, "unrecognised") {
		t.Errorf("a recognised-but-disallowed key should not be called unrecognised; got:\n%s", stderr)
	}
	if !strings.Contains(stderr, GlobalConfigFile) {
		t.Errorf("warning should name the global config; got:\n%s", stderr)
	}
}

// TestIgnoredKeyWarningsAreDeterministic: map iteration order is random, so
// unsorted output would make the warnings shuffle between runs.
func TestIgnoredKeyWarningsAreDeterministic(t *testing.T) {
	clearAssemblaEnv(t)
	configFile := withTempGlobalConfig(t)
	writeFile(t, configFile, "zeta: 1\nalpha: 2\nmiddle: 3\nbeta: 4\n")
	chdir(t, isolatedDir(t))

	first := captureStderr(t, func() { LoadConfig() })
	for i := 0; i < 8; i++ {
		if got := captureStderr(t, func() { LoadConfig() }); got != first {
			t.Fatalf("warning order varies between runs:\n%s\nvs\n%s", first, got)
		}
	}

	// Sorted, so alpha precedes zeta.
	if strings.Index(first, "alpha") > strings.Index(first, "zeta") {
		t.Errorf("keys should be reported in sorted order; got:\n%s", first)
	}
}

// TestWarningsGoToStderr keeps --json output pipeable.
func TestWarningsGoToStderr(t *testing.T) {
	clearAssemblaEnv(t)
	configFile := withTempGlobalConfig(t)
	writeFile(t, configFile, "debug: true\n")
	chdir(t, isolatedDir(t))

	if stderr := captureStderr(t, func() { LoadConfig() }); !strings.Contains(stderr, "debug") {
		t.Errorf("warning should be on stderr; got:\n%s", stderr)
	}
}

// TestAPIURLIsGlobalOnly states the policy in one place: a cloned repository does
// not get to choose where credentials are sent.
func TestAPIURLIsGlobalOnly(t *testing.T) {
	if !globalConfigKeys["api_url"] {
		t.Error("the user's own global config should be able to set api_url")
	}
	if projectConfigKeys["api_url"] {
		t.Error("project config must not be able to set api_url")
	}
}

// TestAllowlistCoversEveryReadKey keeps the allowlist honest: every key the rest
// of the code reads from the config is either declared file-settable or is a
// loader-internal marker.
func TestAllowlistCoversEveryReadKey(t *testing.T) {
	// Mirrors the keys read in cmd/ and internal/.
	readKeys := []string{"api_key", "api_secret", "space", "api_url"}
	for _, key := range readKeys {
		if !globalConfigKeys[key] {
			t.Errorf("%q is read by the CLI but not settable in the global config", key)
		}
	}

	// The env table must not offer a key that files may not set, or precedence
	// would be inconsistent between the two routes.
	for _, key := range []string{"api_key", "api_secret", "space", "api_url"} {
		if !projectConfigKeys[key] && !globalConfigKeys[key] {
			t.Errorf("%q is settable by environment but by no config file", key)
		}
	}
}
