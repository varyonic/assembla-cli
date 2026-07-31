package cmd

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Indirection for tests, which have no terminal to drive.
var (
	stdinFd            = func() int { return int(os.Stdin.Fd()) }
	stdinIsTerminal    = func() bool { return term.IsTerminal(stdinFd()) }
	readPasswordFromFd = term.ReadPassword
)

// readLine reads one line from stdin.
//
// It reads a byte at a time rather than using a buffered reader: bufio reads
// ahead, and anything it swallowed would be lost to the unbuffered read that
// collects the API secret.
func readLine() (string, error) {
	var line []byte
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				break
			}
			line = append(line, buf[0])
		}
		if err != nil {
			if len(line) == 0 {
				return "", err
			}
			break
		}
	}
	return strings.TrimSpace(string(line)), nil
}

// readSecret reads one line from stdin without echoing it to the terminal.
//
// When stdin is not a terminal — a pipe, a script, a here-doc — there is no echo
// to suppress and no terminal to put into raw mode, so it falls back to a plain
// read and scripted logins keep working.
func readSecret() (string, error) {
	if !stdinIsTerminal() {
		return readLine()
	}

	data, err := readPasswordFromFd(stdinFd())
	// ReadPassword consumes the newline without echoing it, leaving the cursor at
	// the end of the prompt.
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
