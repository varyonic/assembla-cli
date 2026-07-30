package internal

import (
	"strings"
	"testing"
)

// allSources is every source api_url can now come from. Project config is absent
// by design — see projectConfigKeys in config.go.
var allSources = []string{SourceDefault, SourceGlobal, SourceEnv}

// TestResolveAPIURLRejectsInsecureOrMalformed covers the checks that apply to
// every source.
func TestResolveAPIURLRejectsInsecureOrMalformed(t *testing.T) {
	cases := []struct {
		name    string
		apiURL  string
		wantErr string
	}{
		{"plain http", "http://api.assembla.com", "must use https"},
		{"http to attacker", "http://127.0.0.1:9000", "must use https"},
		{"uppercase scheme", "HTTP://api.assembla.com", "must use https"},
		{"no scheme", "api.assembla.com", "must use https"},
		{"other scheme", "ftp://api.assembla.com", "must use https"},
		{"empty", "", "empty"},
		{"whitespace only", "   ", "empty"},
		{"no host", "https://", "no host"},
		{"embedded credentials", "https://user:pass@evil.example.com", "must not embed credentials"},
	}

	for _, source := range allSources {
		for _, tc := range cases {
			t.Run(tc.name+"/"+source, func(t *testing.T) {
				withTempGlobalConfig(t)

				got, err := ResolveAPIURL(tc.apiURL, source)
				if err == nil {
					t.Fatalf("expected rejection of %q, got %q", tc.apiURL, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error %q does not mention %q", err, tc.wantErr)
				}
			})
		}
	}
}

func TestResolveAPIURLAcceptsTheDefaultHost(t *testing.T) {
	for _, source := range allSources {
		t.Run(source, func(t *testing.T) {
			withTempGlobalConfig(t)

			got, err := ResolveAPIURL(DefaultAPIURL, source)
			if err != nil {
				t.Fatalf("default host rejected for source %s: %v", source, err)
			}
			if got != DefaultAPIURL {
				t.Errorf("got %q, want %q", got, DefaultAPIURL)
			}
		})
	}
}

// TestResolveAPIURLAcceptsOtherHostsWithoutCeremony: every remaining source is
// one the user controls directly, so an on-prem endpoint needs no confirmation.
func TestResolveAPIURLAcceptsOtherHostsWithoutCeremony(t *testing.T) {
	for _, apiURL := range []string{
		"https://assembla.internal.example.com",
		"https://api.assembla.com:8443",
		"https://internal.example.com/assembla",
	} {
		for _, source := range allSources {
			t.Run(apiURL+"/"+source, func(t *testing.T) {
				withTempGlobalConfig(t)

				got, err := ResolveAPIURL(apiURL, source)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != apiURL {
					t.Errorf("got %q, want %q", got, apiURL)
				}
			})
		}
	}
}

func TestResolveAPIURLTrimsSurroundingWhitespace(t *testing.T) {
	withTempGlobalConfig(t)

	got, err := ResolveAPIURL("  "+DefaultAPIURL+"  ", SourceGlobal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != DefaultAPIURL {
		t.Errorf("got %q, want %q", got, DefaultAPIURL)
	}
}

// TestResolveAPIURLErrorNamesTheSource: when a bad value is rejected, the message
// has to say which file or variable to go and fix.
func TestResolveAPIURLErrorNamesTheSource(t *testing.T) {
	configFile := withTempGlobalConfig(t)

	tests := []struct {
		source string
		want   string
	}{
		{SourceGlobal, configFile},
		{SourceEnv, "ASSEMBLA_API_URL"},
		{SourceDefault, "built-in default"},
	}

	for _, tc := range tests {
		t.Run(tc.source, func(t *testing.T) {
			_, err := ResolveAPIURL("http://example.com", tc.source)
			if err == nil {
				t.Fatal("expected rejection")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should mention %q", err, tc.want)
			}
		})
	}
}
