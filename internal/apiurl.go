package internal

import (
	"fmt"
	"net/url"
	"strings"
)

// DefaultAPIURL is the Assembla SaaS API endpoint.
const DefaultAPIURL = "https://api.assembla.com"

// Where a config value came from, used to say so in diagnostics.
//
// There is no project source for api_url: a .assembla.yml can arrive as part of a
// cloned repository, so it is not allowed to choose where credentials are sent.
// See projectConfigKeys in config.go.
const (
	SourceDefault = "default"
	SourceGlobal  = "global"
	SourceEnv     = "env"
)

// ResolveAPIURL validates apiURL before any credentials are sent to it.
//
// Every accepted source (the built-in default, the user's global config, the
// environment) is one the user controls directly, so the only question left is
// whether the URL itself is safe to send credentials to.
func ResolveAPIURL(apiURL, source string) (string, error) {
	where := describeSource(source)

	trimmed := strings.TrimSpace(apiURL)
	if trimmed == "" {
		return "", fmt.Errorf("api_url is empty (%s)", where)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid api_url %q (%s): %w", trimmed, where, err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("api_url must use https, got %q (%s); "+
			"credentials travel as request headers and would otherwise cross the network in cleartext",
			trimmed, where)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("api_url %q has no host (%s)", trimmed, where)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("api_url %q must not embed credentials (%s)", trimmed, where)
	}

	return trimmed, nil
}

func describeSource(source string) string {
	switch source {
	case SourceGlobal:
		return "configured in " + GlobalConfigFile
	case SourceEnv:
		return "set by ASSEMBLA_API_URL"
	default:
		return "built-in default"
	}
}
