package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// stdinFrom points os.Stdin at a file containing content.
func stdinFrom(t *testing.T, content string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open stdin: %v", err)
	}

	prev := os.Stdin
	os.Stdin = file
	t.Cleanup(func() {
		os.Stdin = prev
		file.Close()
	})
}

// stubTerminal forces the terminal check and password reader for a test.
func stubTerminal(t *testing.T, isTerminal bool, password string, readErr error) *int {
	t.Helper()

	calls := 0
	prevIsTerm, prevRead := stdinIsTerminal, readPasswordFromFd
	stdinIsTerminal = func() bool { return isTerminal }
	readPasswordFromFd = func(int) ([]byte, error) {
		calls++
		return []byte(password), readErr
	}
	t.Cleanup(func() {
		stdinIsTerminal, readPasswordFromFd = prevIsTerm, prevRead
	})
	return &calls
}

func TestReadLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain line", "hello\n", "hello"},
		{"windows line ending", "hello\r\n", "hello"},
		{"surrounding whitespace", "  hello  \n", "hello"},
		{"empty line", "\n", ""},
		{"no trailing newline", "hello", "hello"},
		{"stops at first newline", "first\nsecond\n", "first"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdinFrom(t, tc.input)

			got, err := readLine()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestReadLineDoesNotReadAhead is the reason readLine avoids bufio: anything
// buffered past the newline would be lost to the next read, which is how the API
// secret would go missing.
func TestReadLineDoesNotReadAhead(t *testing.T) {
	stdinFrom(t, "the-key\nthe-secret\nmyspace\n")

	first, err := readLine()
	if err != nil {
		t.Fatalf("first read: %v", err)
	}
	second, err := readLine()
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	third, err := readLine()
	if err != nil {
		t.Fatalf("third read: %v", err)
	}

	if first != "the-key" || second != "the-secret" || third != "myspace" {
		t.Errorf("got %q, %q, %q; want the three lines in order", first, second, third)
	}
}

func TestReadLineReturnsErrorOnEmptyInput(t *testing.T) {
	stdinFrom(t, "")

	if _, err := readLine(); err == nil {
		t.Error("expected an error at immediate EOF")
	}
}

// TestReadSecretUsesNoEchoReaderOnTerminal: the whole point of the fix.
func TestReadSecretUsesNoEchoReaderOnTerminal(t *testing.T) {
	calls := stubTerminal(t, true, "  s3cret  ", nil)
	// Provide stdin content too: if readSecret wrongly fell back to readLine, it
	// would return this instead and the test would notice.
	stdinFrom(t, "echoed-from-stdin\n")

	got, err := readSecret()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "s3cret" {
		t.Errorf("got %q, want the no-echo reader's value (trimmed)", got)
	}
	if *calls != 1 {
		t.Errorf("no-echo reader called %d times, want 1", *calls)
	}
}

// TestReadSecretFallsBackWhenNotATerminal keeps scripted logins working.
func TestReadSecretFallsBackWhenNotATerminal(t *testing.T) {
	calls := stubTerminal(t, false, "unused", nil)
	stdinFrom(t, "piped-secret\n")

	got, err := readSecret()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "piped-secret" {
		t.Errorf("got %q, want the piped value", got)
	}
	if *calls != 0 {
		t.Errorf("no-echo reader called %d times, want 0 when stdin is not a terminal", *calls)
	}
}

func TestReadSecretPropagatesReaderError(t *testing.T) {
	wantErr := errors.New("ioctl failed")
	stubTerminal(t, true, "", wantErr)
	stdinFrom(t, "")

	if _, err := readSecret(); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}
