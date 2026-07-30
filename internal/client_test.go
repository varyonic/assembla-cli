package internal

import (
	"net/http"
	"net/url"
	"testing"
)

// credentialedRequest builds a request carrying the API credential headers.
func credentialedRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse %s: %v", rawURL, err)
	}
	req := &http.Request{URL: parsed, Header: http.Header{}}
	req.Header.Set("X-Api-Key", "KEY")
	req.Header.Set("X-Api-Secret", "SECRET")
	return req
}

// TestRedirectKeepsCredentialsOnSameHost: the API is allowed to redirect
// internally without breaking authentication.
func TestRedirectKeepsCredentialsOnSameHost(t *testing.T) {
	original := credentialedRequest(t, "https://api.assembla.com/v1/user")

	for _, target := range []string{
		"https://api.assembla.com/v1/user/",
		"https://api.assembla.com/v1/other",
		"https://API.ASSEMBLA.COM/v1/user", // host comparison ignores case
	} {
		next := credentialedRequest(t, target)
		if err := stripCredentialsOnHostChange(next, []*http.Request{original}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next.Header.Get("X-Api-Key") != "KEY" || next.Header.Get("X-Api-Secret") != "SECRET" {
			t.Errorf("same-host redirect to %s should keep credentials", target)
		}
	}
}

// TestRedirectStripsCredentialsOnHostChange: Go forwards custom headers across
// redirects and only strips the ones it recognises, so a validated host could
// otherwise hand the credentials to a third party.
func TestRedirectStripsCredentialsOnHostChange(t *testing.T) {
	original := credentialedRequest(t, "https://api.assembla.com/v1/user")

	for _, target := range []string{
		"https://evil.example.com/collect",
		"http://evil.example.com/collect",
		"https://api.assembla.com.evil.example.com/collect",
		"https://api.assembla.com:8443/v1/user", // port change is a different endpoint
	} {
		next := credentialedRequest(t, target)
		if err := stripCredentialsOnHostChange(next, []*http.Request{original}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := next.Header.Get("X-Api-Key"); got != "" {
			t.Errorf("redirect to %s leaked X-Api-Key (%q)", target, got)
		}
		if got := next.Header.Get("X-Api-Secret"); got != "" {
			t.Errorf("redirect to %s leaked X-Api-Secret (%q)", target, got)
		}
	}
}

// TestRedirectStripsAcrossMultipleHops: the comparison is against the original
// request, so a chain cannot launder the host one hop at a time.
func TestRedirectStripsAcrossMultipleHops(t *testing.T) {
	original := credentialedRequest(t, "https://api.assembla.com/v1/user")
	hop := credentialedRequest(t, "https://api.assembla.com/v1/redirect")
	final := credentialedRequest(t, "https://evil.example.com/collect")

	if err := stripCredentialsOnHostChange(final, []*http.Request{original, hop}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if final.Header.Get("X-Api-Key") != "" || final.Header.Get("X-Api-Secret") != "" {
		t.Error("multi-hop redirect leaked credentials to a different host")
	}
}

func TestRedirectLimitIsEnforced(t *testing.T) {
	original := credentialedRequest(t, "https://api.assembla.com/v1/user")

	var via []*http.Request
	for i := 0; i < 10; i++ {
		via = append(via, original)
	}

	next := credentialedRequest(t, "https://api.assembla.com/v1/user")
	if err := stripCredentialsOnHostChange(next, via); err == nil {
		t.Error("expected an error once the redirect limit is reached")
	}
}

// TestNewAssemblaClientInstallsRedirectGuard: the protection is worthless if the
// client is ever constructed without it.
func TestNewAssemblaClientInstallsRedirectGuard(t *testing.T) {
	client := NewAssemblaClient("KEY", "SECRET", DefaultAPIURL)

	if client.httpClient.CheckRedirect == nil {
		t.Fatal("client must install a CheckRedirect guard")
	}

	next := credentialedRequest(t, "https://evil.example.com/collect")
	original := credentialedRequest(t, "https://api.assembla.com/v1/user")
	if err := client.httpClient.CheckRedirect(next, []*http.Request{original}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next.Header.Get("X-Api-Key") != "" {
		t.Error("client redirect guard did not strip credentials")
	}
}

func TestNewAssemblaClientTrimsTrailingSlash(t *testing.T) {
	client := NewAssemblaClient("KEY", "SECRET", "https://api.assembla.com///")
	if client.apiURL != "https://api.assembla.com" {
		t.Errorf("apiURL = %q, want trailing slashes removed", client.apiURL)
	}
}
