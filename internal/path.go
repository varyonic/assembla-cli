package internal

import (
	"net/url"
	"strings"
)

// APIPath builds a request path from segments, escaping each one.
//
// Interpolating a value straight into a path lets it reshape the request: a "?"
// or "#" truncates the path and takes over the query, and "/.." walks to a
// different endpoint. Space names and ticket numbers can come from a project
// .assembla.yml, so they are not necessarily the user's own input.
//
// Segments are the path components without separators:
//
//	APIPath("spaces", space, "tickets") -> "/spaces/my-space/tickets"
func APIPath(segments ...string) string {
	var path strings.Builder
	for _, segment := range segments {
		path.WriteByte('/')
		path.WriteString(url.PathEscape(segment))
	}
	return path.String()
}
