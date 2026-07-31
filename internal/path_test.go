package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAPIPathBuildsOrdinaryPaths(t *testing.T) {
	tests := []struct {
		segments []string
		want     string
	}{
		{[]string{"user"}, "/user"},
		{[]string{"spaces"}, "/spaces"},
		{[]string{"spaces", "my-space", "tickets"}, "/spaces/my-space/tickets"},
		{[]string{"spaces", "my_space", "tickets", "42"}, "/spaces/my_space/tickets/42"},
		{[]string{"spaces", "Space123", "milestones", "upcoming"}, "/spaces/Space123/milestones/upcoming"},
		{[]string{"spaces", "s", "tickets", "7", "ticket_comments"}, "/spaces/s/tickets/7/ticket_comments"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := APIPath(tc.segments...); got != tc.want {
				t.Errorf("APIPath(%q) = %q, want %q", tc.segments, got, tc.want)
			}
		})
	}
}

// TestAPIPathNeutralisesReshapingCharacters covers the characters that let a
// segment take over the request rather than sit inside the path.
func TestAPIPathNeutralisesReshapingCharacters(t *testing.T) {
	for _, segment := range []string{
		"x?per_page=999",
		"x#fragment",
		"../../user",
		"myspace/../../user",
		"a/b",
		"a b",
		"..",
		"/leading",
		"trailing/",
		"%2e%2e",
		"a%2Fb",
	} {
		t.Run(segment, func(t *testing.T) {
			got := APIPath("spaces", segment, "tickets")

			// The segment must not introduce structure of its own.
			for _, forbidden := range []string{"?", "#"} {
				if strings.Contains(got, forbidden) {
					t.Errorf("APIPath produced %q, containing %q", got, forbidden)
				}
			}
			// Exactly three separators means the segment did not add its own.
			if n := strings.Count(got, "/"); n != 3 {
				t.Errorf("APIPath produced %q with %d separators, want 3", got, n)
			}
			if !strings.HasPrefix(got, "/spaces/") || !strings.HasSuffix(got, "/tickets") {
				t.Errorf("APIPath produced %q, want the segment contained between the literals", got)
			}
		})
	}
}

// TestAPIPathRoundTripsThroughURLParsing is the property that matters: whatever
// the segment contains, the parsed request must keep the intended path and query.
func TestAPIPathRoundTripsThroughURLParsing(t *testing.T) {
	for _, segment := range []string{
		"my-space",
		"x?per_page=999",
		"x#fragment",
		"../../user",
		"a b",
	} {
		t.Run(segment, func(t *testing.T) {
			raw := "https://api.assembla.com/v1" + APIPath("spaces", segment, "tickets") + "?page=1"

			parsed, err := url.Parse(raw)
			if err != nil {
				t.Fatalf("parse %q: %v", raw, err)
			}
			if parsed.Host != "api.assembla.com" {
				t.Errorf("host = %q, want api.assembla.com", parsed.Host)
			}
			if parsed.RawQuery != "page=1" {
				t.Errorf("query = %q, want the caller's query intact", parsed.RawQuery)
			}
			if parsed.Fragment != "" {
				t.Errorf("fragment = %q, want none", parsed.Fragment)
			}
			// Path decodes back to the literal segment, in the intended position.
			if want := "/v1/spaces/" + segment + "/tickets"; parsed.Path != want {
				t.Errorf("path = %q, want %q", parsed.Path, want)
			}
		})
	}
}

// TestClientSendsEscapedPath drives a real HTTP server, which is where the
// original defect showed itself: the query was rewritten and /tickets vanished.
func TestClientSendsEscapedPath(t *testing.T) {
	type received struct {
		path  string
		query string
	}

	var got received
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = received{path: r.URL.Path, query: r.URL.RawQuery}
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()

	client := NewAssemblaClient("KEY", "SECRET", server.URL)

	for _, segment := range []string{
		"my-space",
		"x?per_page=999",
		"x#fragment",
		"../../user",
		"a b",
	} {
		t.Run(segment, func(t *testing.T) {
			got = received{}
			path := APIPath("spaces", segment, "tickets")

			if _, err := client.Get(path, map[string]string{"page": "1"}); err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if want := "/v1/spaces/" + segment + "/tickets"; got.path != want {
				t.Errorf("server saw path %q, want %q", got.path, want)
			}
			if got.query != "page=1" {
				t.Errorf("server saw query %q, want page=1", got.query)
			}
		})
	}
}

// TestClientSendsSeparatorsEncodedOnTheWire: a server routes on the escaped path,
// so the separators inside a segment must arrive as %2F rather than as real ones.
func TestClientSendsSeparatorsEncodedOnTheWire(t *testing.T) {
	var escaped, segments string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escaped = r.URL.EscapedPath()
		segments = fmt.Sprint(len(strings.Split(strings.Trim(r.URL.EscapedPath(), "/"), "/")))
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()

	client := NewAssemblaClient("KEY", "SECRET", server.URL)

	if _, err := client.Get(APIPath("spaces", "../../user", "tickets"), nil); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !strings.Contains(escaped, "%2F") {
		t.Errorf("wire path %q should encode the separators as %%2F", escaped)
	}
	if strings.Contains(escaped, "/../") {
		t.Errorf("wire path %q still contains a traversal segment", escaped)
	}
	// /v1, /spaces, /<segment>, /tickets — the segment stayed one component.
	if segments != "4" {
		t.Errorf("wire path %q has %s components, want 4", escaped, segments)
	}
}

// TestClientRejectsNothingLegitimate: ordinary identifiers must reach the server
// byte for byte, so escaping has not changed normal behaviour.
func TestClientRejectsNothingLegitimate(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()

	client := NewAssemblaClient("KEY", "SECRET", server.URL)

	for _, tc := range []struct{ space, number, want string }{
		{"myspace", "42", "/v1/spaces/myspace/tickets/42"},
		{"my-space", "1", "/v1/spaces/my-space/tickets/1"},
		{"my_space_2", "99999", "/v1/spaces/my_space_2/tickets/99999"},
		{"Space.Name", "7", "/v1/spaces/Space.Name/tickets/7"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			gotPath = ""
			if _, err := client.Get(APIPath("spaces", tc.space, "tickets", tc.number), nil); err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if gotPath != tc.want {
				t.Errorf("server saw %q, want %q", gotPath, tc.want)
			}
		})
	}
}
