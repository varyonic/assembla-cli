package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintTable prints a slice of maps as a formatted table.
func PrintTable(rows []interface{}, columns []string, headers []string) {
	if len(rows) == 0 {
		fmt.Println("No results.")
		return
	}

	if headers == nil {
		headers = columns
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print separator
	var sep []string
	for _, h := range headers {
		sep = append(sep, strings.Repeat("-", len(h)))
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))

	// Print rows
	for _, row := range rows {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		var vals []string
		for _, col := range columns {
			val := rowMap[col]
			if val == nil {
				vals = append(vals, "")
			} else {
				vals = append(vals, fmt.Sprintf("%v", val))
			}
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}

	w.Flush()
}

// PrintJSON prints data as indented JSON.
func PrintJSON(data interface{}) {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(out))
}

// PrintDetail prints a single record's fields with aligned keys.
func PrintDetail(data interface{}, fields []string) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		fmt.Println("No data.")
		return
	}

	maxKey := 0
	for _, f := range fields {
		if len(f) > maxKey {
			maxKey = len(f)
		}
	}

	for _, field := range fields {
		val := dataMap[field]
		if val == nil {
			val = ""
		}
		fmt.Printf("  %-*s  %v\n", maxKey, field, val)
	}
}

// ToSlice converts an interface{} that is expected to be []interface{} to that type.
// Returns nil if the input is nil or not a slice.
func ToSlice(data interface{}) []interface{} {
	if data == nil {
		return nil
	}
	if arr, ok := data.([]interface{}); ok {
		return arr
	}
	return nil
}
